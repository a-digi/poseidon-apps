package repko

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/coder/websocket"
)

type wsHandler struct {
	mgr *Manager
}

func newWSHandler(mgr *Manager) *wsHandler {
	return &wsHandler{mgr: mgr}
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		log.Printf("ws: accept failed (remote=%s err=%v)", r.RemoteAddr, err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "")

	ctx := r.Context()
	log.Printf("ws: connection opened (remote=%s)", r.RemoteAddr)

	_, raw, err := c.Read(ctx)
	if err != nil {
		return
	}
	var hello Hello
	if json.Unmarshal(raw, &hello) != nil || hello.Type != MsgHello {
		writeErrorFrame(ctx, c, "expected hello")
		return
	}
	log.Printf("ws: hello (remote=%s room=%q name=%q resume=%t)", r.RemoteAddr, hello.Room, hello.Name, hello.ResumeToken != "")
	if strings.TrimSpace(hello.Room) == "" {
		log.Printf("ws: rejected hello (remote=%s reason=empty_room)", r.RemoteAddr)
		writeErrorFrame(ctx, c, "room is required")
		return
	}

	var (
		room   *Room
		player *Player
	)

	if hello.ResumeToken != "" {
		var ok bool
		room, ok = h.mgr.GetRoom(hello.Room)
		if !ok {
			log.Printf("ws: resume rejected (room=%q reason=unknown_room)", hello.Room)
			writeErrorFrame(ctx, c, "resume failed: unknown room")
			return
		}

		playerID, found := room.playerIDByResumeToken(hello.ResumeToken)
		if !found {
			log.Printf("ws: resume rejected (room=%s reason=token_unknown)", room.Code)
			writeErrorFrame(ctx, c, "resume failed: token unknown")
			return
		}

		room.mu.Lock()
		if room.closed {
			room.mu.Unlock()
			log.Printf("ws: resume rejected (room=%s reason=room_closed)", room.Code)
			writeErrorFrame(ctx, c, "resume failed: room not started")
			return
		}
		if room.state == nil {
			room.mu.Unlock()
			log.Printf("ws: resume rejected (room=%s reason=lobby)", room.Code)
			writeErrorFrame(ctx, c, "resume failed: room not started")
			return
		}
		if room.state.Phase == PhaseGameOver {
			room.mu.Unlock()
			log.Printf("ws: resume rejected (room=%s reason=game_over)", room.Code)
			writeErrorFrame(ctx, c, "resume failed: game over")
			return
		}
		ps := room.state.playerByID(playerID)
		if ps == nil {
			room.mu.Unlock()
			log.Printf("ws: resume rejected (room=%s reason=player_gone)", room.Code)
			writeErrorFrame(ctx, c, "resume failed: player gone")
			return
		}
		if room.playerByIDLocked(playerID) != nil {
			room.mu.Unlock()
			log.Printf("ws: resume rejected (room=%s player_id=%s reason=already_connected)", room.Code, playerID)
			writeErrorFrame(ctx, c, "resume failed: already connected")
			return
		}

		player = &Player{
			ID:     playerID,
			Name:   ps.Name,
			Color:  ps.Color,
			SendCh: make(chan []byte, sendChCapacity),
		}
		room.players = append(room.players, player)
		room.LastEmptyAt = nil
		ps.Disconnected = false
		room.mu.Unlock()

		log.Printf("ws: resume accepted (room=%s player_id=%s)", room.Code, player.ID)

		welcome := Welcome{
			Type:        MsgWelcome,
			PlayerID:    player.ID,
			Room:        room.Code,
			ResumeToken: hello.ResumeToken,
			You:         WelcomeYou{Name: player.Name, Color: player.Color},
		}
		welcomeBytes, _ := json.Marshal(welcome)
		if err := c.Write(ctx, websocket.MessageText, welcomeBytes); err != nil {
			room.Remove(player.ID)
			return
		}
		log.Printf("ws: welcome sent (remote=%s player_id=%s room=%s mode=resume)", r.RemoteAddr, player.ID, room.Code)
		room.Broadcast()
	} else {
		trimmed := strings.TrimSpace(hello.Name)
		if len(trimmed) < 1 || len(trimmed) > 32 {
			log.Printf("ws: rejected hello (remote=%s reason=invalid_name name_len=%d)", r.RemoteAddr, len(trimmed))
			writeErrorFrame(ctx, c, "name must be 1-32 characters")
			return
		}

		var ok bool
		room, ok = h.mgr.GetRoom(hello.Room)
		if !ok {
			log.Printf("ws: rejected hello (remote=%s room=%q reason=unknown_room)", r.RemoteAddr, hello.Room)
			writeErrorFrame(ctx, c, "unknown room")
			return
		}

		player = &Player{
			ID:     newPlayerID(),
			Name:   trimmed,
			SendCh: make(chan []byte, sendChCapacity),
		}
		if !room.Add(player) {
			if room.PlayerCount() >= maxPlayers {
				writeErrorFrame(ctx, c, "room is full")
			} else {
				writeErrorFrame(ctx, c, "room is closed")
			}
			return
		}
		log.Printf("game: player joined (room=%s player_id=%s name=%q color=%s)", room.Code, player.ID, player.Name, player.Color)

		token := room.generateResumeToken(player.ID)
		welcome := Welcome{
			Type:        MsgWelcome,
			PlayerID:    player.ID,
			Room:        room.Code,
			ResumeToken: token,
			You:         WelcomeYou{Name: player.Name, Color: player.Color},
		}
		welcomeBytes, _ := json.Marshal(welcome)
		if err := c.Write(ctx, websocket.MessageText, welcomeBytes); err != nil {
			room.Remove(player.ID)
			return
		}
		log.Printf("ws: welcome sent (remote=%s player_id=%s room=%s)", r.RemoteAddr, player.ID, room.Code)

		room.Broadcast()
		if room.ShouldAutoStart() {
			log.Printf("game: auto-starting room=%s (expected players reached)", room.Code)
			_ = room.StartGame()
			room.mu.Lock()
			room.armTurnTimerLocked()
			room.mu.Unlock()
		}
	}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-player.SendCh:
				if !ok {
					return
				}
				if err := c.Write(ctx, websocket.MessageText, msg); err != nil {
					return
				}
			}
		}
	}()

	var readErr error
readLoop:
	for {
		_, raw, err := c.Read(ctx)
		if err != nil {
			readErr = err
			break
		}
		var head struct {
			Type MessageType `json:"type"`
		}
		if json.Unmarshal(raw, &head) != nil {
			sendError(player, "invalid message")
			continue
		}
		switch head.Type {
		case MsgHello:
			continue
		case MsgPickCivilization:
			var m PickCivilization
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionPickCivilization{CivilizationID: m.CivilizationID})
		case MsgPickStartingTile:
			var m PickStartingTile
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionPickStartingTile{Q: m.Q, R: m.R})
		case MsgRecruit:
			var m Recruit
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionRecruit{Q: m.Q, R: m.R, Unit: m.Unit, Count: m.Count})
		case MsgUpgrade:
			var m Upgrade
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionUpgrade{Q: m.Q, R: m.R, StackIndex: m.StackIndex})
		case MsgMove:
			var m Move
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionMove{FromQ: m.FromQ, FromR: m.FromR, ToQ: m.ToQ, ToR: m.ToR, Units: m.Units})
		case MsgAttack:
			var m Attack
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionAttack{FromQ: m.FromQ, FromR: m.FromR, ToQ: m.ToQ, ToR: m.ToR, Units: m.Units})
		case MsgBuyTile:
			var m BuyTile
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionBuyTile{Q: m.Q, R: m.R})
		case MsgUpgradeTile:
			var m UpgradeTileMsg
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionUpgradeTile{Q: m.Q, R: m.R})
		case MsgOfferDiplomacy:
			var m OfferDiplomacy
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionOfferDiplomacy{Q: m.Q, R: m.R})
		case MsgAcceptDiplomacy:
			var m AcceptDiplomacy
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionAcceptDiplomacy{Q: m.Q, R: m.R})
		case MsgDeclineDiplomacy:
			var m DeclineDiplomacy
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionDeclineDiplomacy{Q: m.Q, R: m.R})
		case MsgCancelDiplomacy:
			var m CancelDiplomacy
			if json.Unmarshal(raw, &m) != nil {
				sendError(player, "invalid message")
				continue
			}
			h.dispatch(room, player, ActionCancelDiplomacy{Q: m.Q, R: m.R})
		case MsgEndTurn:
			h.dispatch(room, player, ActionEndTurn{})
		case MsgLeaveGame:
			room.LeaveGame(player.ID)
			log.Printf("game: repko player left (room=%s player_id=%s)", room.Code, player.ID)
			break readLoop
		default:
			sendError(player, "unknown message type")
		}
	}

	room.Remove(player.ID)
	player.CloseSend()
	<-writerDone
	c.Close(websocket.StatusNormalClosure, "")
	if readErr != nil {
		log.Printf("ws: disconnected (remote=%s player_id=%s room=%s err=%v)", r.RemoteAddr, player.ID, room.Code, readErr)
	} else {
		log.Printf("ws: disconnected (remote=%s player_id=%s room=%s)", r.RemoteAddr, player.ID, room.Code)
	}
}

// dispatch runs an action against the room state under r.mu. On accept,
// broadcasts the new state. On reject, returns a private error frame to the
// originating player without broadcasting.
func (h *wsHandler) dispatch(room *Room, player *Player, action any) {
	actionType := actionTypeName(action)
	room.mu.Lock()
	state := room.state
	if state == nil {
		room.mu.Unlock()
		log.Printf("game: repko action rejected (room=%s player_id=%s action=%s reason=no_game)", room.Code, player.ID, actionType)
		sendError(player, ErrNoGame.Error())
		return
	}
	err := ValidateAndApply(state, player.ID, action)
	phase := state.Phase
	if err == nil {
		room.armTurnTimerLocked()
	}
	room.mu.Unlock()
	if err != nil {
		log.Printf("game: repko action rejected (room=%s player_id=%s action=%s phase=%s reason=%v)", room.Code, player.ID, actionType, phase, err)
		sendError(player, err.Error())
		return
	}
	log.Printf("game: repko action accepted (room=%s player_id=%s action=%s phase=%s)", room.Code, player.ID, actionType, phase)
	room.Broadcast()
}

func actionTypeName(action any) string {
	switch action.(type) {
	case ActionPickCivilization:
		return "pick_civilization"
	case ActionPickStartingTile:
		return "pick_starting_tile"
	case ActionRecruit:
		return "recruit"
	case ActionUpgrade:
		return "upgrade"
	case ActionMove:
		return "move"
	case ActionAttack:
		return "attack"
	case ActionBuyTile:
		return "buy_tile"
	case ActionUpgradeTile:
		return "upgrade_tile"
	case ActionOfferDiplomacy:
		return "offer_diplomacy"
	case ActionAcceptDiplomacy:
		return "accept_diplomacy"
	case ActionDeclineDiplomacy:
		return "decline_diplomacy"
	case ActionCancelDiplomacy:
		return "cancel_diplomacy"
	case ActionEndTurn:
		return "end_turn"
	default:
		return "unknown"
	}
}

func sendError(p *Player, msg string) {
	b, _ := json.Marshal(ErrorMsg{Type: MsgError, Message: msg})
	select {
	case p.SendCh <- b:
	default:
	}
}

// writeErrorFrame writes a typed error directly to the websocket conn, used
// during the pre-Add handshake before a SendCh exists.
func writeErrorFrame(ctx context.Context, c *websocket.Conn, msg string) {
	b, _ := json.Marshal(ErrorMsg{Type: MsgError, Message: msg})
	_ = c.Write(ctx, websocket.MessageText, b)
}

func newPlayerID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}
