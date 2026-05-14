package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"coco-gg-plugin/game"

	"github.com/coder/websocket"
)

type WSHandler struct {
	mgr *game.Manager
}

func NewWSHandler(mgr *game.Manager) *WSHandler {
	return &WSHandler{mgr: mgr}
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// InsecureSkipVerify: the host's reverse proxy strips the Origin header
	// before forwarding, so origin checks here would always fail. Auth is
	// enforced at the host edge by the mobile-token middleware.
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
	var hello game.Hello
	if json.Unmarshal(raw, &hello) != nil || hello.Type != game.MsgHello {
		writeError(ctx, c, "expected hello")
		return
	}
	log.Printf("ws: hello (remote=%s room=%q name=%q)", r.RemoteAddr, hello.Room, hello.Name)
	trimmed := strings.TrimSpace(hello.Name)
	if len(trimmed) < 1 || len(trimmed) > 32 {
		log.Printf("ws: rejected hello (remote=%s reason=invalid_name name_len=%d)", r.RemoteAddr, len(trimmed))
		writeError(ctx, c, "name must be 1-32 characters")
		return
	}
	if strings.TrimSpace(hello.Room) == "" {
		log.Printf("ws: rejected hello (remote=%s reason=empty_room)", r.RemoteAddr)
		writeError(ctx, c, "room is required")
		return
	}

	room, player, err := h.mgr.Join(hello.Room, trimmed)
	if errors.Is(err, game.ErrUnknownRoom) {
		log.Printf("ws: rejected hello (remote=%s room=%q reason=unknown_room)", r.RemoteAddr, hello.Room)
		writeError(ctx, c, "unknown room")
		return
	}
	if err != nil {
		writeError(ctx, c, err.Error())
		return
	}

	welcome := game.Welcome{
		Type:     game.MsgWelcome,
		PlayerID: player.ID,
		Room:     room.Code,
		TickHz:   20,
		Arena:    room.Arena,
	}
	welcomeBytes, _ := json.Marshal(welcome)
	if err := c.Write(ctx, websocket.MessageText, welcomeBytes); err != nil {
		h.mgr.Leave(room.Code, player.ID)
		return
	}
	log.Printf("ws: welcome sent (remote=%s player_id=%s room=%s)", r.RemoteAddr, player.ID, room.Code)

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
	for {
		_, raw, err := c.Read(ctx)
		if err != nil {
			readErr = err
			break
		}
		var in game.Input
		if json.Unmarshal(raw, &in) != nil || in.Type != game.MsgInput {
			continue
		}
		room.SetInput(player.ID, in)
	}

	h.mgr.Leave(room.Code, player.ID)
	player.CloseSend()
	<-writerDone
	c.Close(websocket.StatusNormalClosure, "")
	if readErr != nil {
		log.Printf("ws: disconnected (remote=%s player_id=%s room=%s err=%v)", r.RemoteAddr, player.ID, room.Code, readErr)
	} else {
		log.Printf("ws: disconnected (remote=%s player_id=%s room=%s)", r.RemoteAddr, player.ID, room.Code)
	}
}

func writeError(ctx context.Context, c *websocket.Conn, msg string) {
	b, _ := json.Marshal(game.ErrorMsg{Type: game.MsgError, Message: msg})
	_ = c.Write(ctx, websocket.MessageText, b)
}
