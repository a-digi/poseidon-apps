package repko

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

const (
	sendChCapacity = 16
	maxPlayers     = 6
)

var colorPalette = []string{
	"#dc2626", "#2563eb", "#16a34a", "#ea580c",
	"#9333ea", "#0d9488", "#db2777", "#ca8a04",
}

type Player struct {
	ID            string
	Name          string
	Color         string
	SendCh        chan []byte
	closeSendOnce sync.Once
}

func (p *Player) CloseSend() {
	p.closeSendOnce.Do(func() { close(p.SendCh) })
}

type Room struct {
	Code        string
	CreatedAt   time.Time
	LastEmptyAt *time.Time
	mu          sync.Mutex
	closed      bool
	players     []*Player
	state       *GameState
}

func NewRoom(code string) *Room {
	now := time.Now()
	return &Room{
		Code:        code,
		CreatedAt:   now,
		LastEmptyAt: &now,
		players:     make([]*Player, 0, maxPlayers),
	}
}

func (r *Room) Phase() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.phaseLocked()
}

func (r *Room) phaseLocked() string {
	if r.state == nil {
		return string(PhaseLobby)
	}
	return string(r.state.Phase)
}

func (r *Room) PlayerCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.players)
}

func (r *Room) Digest() RoomDigest {
	r.mu.Lock()
	defer r.mu.Unlock()
	players := make([]PlayerDigest, 0, len(r.players))
	for _, p := range r.players {
		players = append(players, PlayerDigest{ID: p.ID, Name: p.Name})
	}
	return RoomDigest{
		Code:      r.Code,
		Phase:     r.phaseLocked(),
		Players:   players,
		CreatedAt: r.CreatedAt.Unix(),
	}
}

// Add inserts a player and assigns a color by join index in the palette.
// Returns false if the room is closed or full.
func (r *Room) Add(p *Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return false
	}
	if len(r.players) >= maxPlayers {
		return false
	}
	p.Color = colorPalette[len(r.players)%len(colorPalette)]
	wasEmpty := len(r.players) == 0
	r.players = append(r.players, p)
	if wasEmpty {
		r.LastEmptyAt = nil
	}
	return true
}

// Remove pulls a player out and broadcasts the resulting state to remaining players.
// If a game is in progress, marks the player Disconnected.
func (r *Room) Remove(playerID string) {
	r.mu.Lock()
	if !r.removeLocked(playerID) || r.closed {
		r.mu.Unlock()
		return
	}
	if r.state != nil && r.state.Phase != PhaseLobby && r.state.Phase != PhaseGameOver {
		if ps := r.state.playerByID(playerID); ps != nil {
			ps.Disconnected = true
		}
		r.handleDisconnectLocked(playerID)
	}
	if len(r.players) == 0 {
		now := time.Now()
		r.LastEmptyAt = &now
	}
	r.mu.Unlock()
	r.Broadcast()
}

func (r *Room) handleDisconnectLocked(playerID string) {
	connected := 0
	for _, ps := range r.state.Players {
		if !ps.Disconnected {
			connected++
		}
	}
	if connected < 1 {
		r.state.Phase = PhaseGameOver
		r.state.Current = nil
		r.state.WinnerID = ""
		return
	}
	wasCurrent := r.state.Current != nil && r.state.Current.PlayerID == playerID
	switch r.state.Phase {
	case PhaseCivPick:
		if wasCurrent {
			advanceCivPick(r.state)
		}
	case PhaseTilePick:
		if wasCurrent {
			advanceTilePick(r.state)
		}
	case PhasePlaying:
		if wasCurrent {
			advancePlayingTurn(r.state)
		}
	}
	cleanupStaleDiplomacy(r.state)
	recomputeCounters(r.state)
}

func (r *Room) removeLocked(playerID string) bool {
	for i, p := range r.players {
		if p.ID == playerID {
			r.players = append(r.players[:i], r.players[i+1:]...)
			return true
		}
	}
	return false
}

// Kick sends an error frame, closes the player's send channel, and removes them.
func (r *Room) Kick(playerID, reason string) bool {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return false
	}
	var target *Player
	for _, p := range r.players {
		if p.ID == playerID {
			target = p
			break
		}
	}
	if target == nil {
		r.mu.Unlock()
		return false
	}
	errMsg, _ := json.Marshal(ErrorMsg{Type: MsgError, Message: reason})
	select {
	case target.SendCh <- errMsg:
	default:
	}
	r.mu.Unlock()

	target.CloseSend()
	r.Remove(playerID)
	return true
}

// StartGame transitions the room from lobby to civ_pick.
func (r *Room) StartGame() error {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		log.Printf("game: repko start rejected (room=%s reason=closed)", r.Code)
		return ErrAlreadyStarted
	}
	if r.state != nil && r.state.Phase != PhaseLobby {
		phase := r.state.Phase
		code := r.Code
		r.mu.Unlock()
		log.Printf("game: repko start rejected (room=%s reason=already_started phase=%s)", code, phase)
		return ErrAlreadyStarted
	}
	if len(r.players) < 2 {
		count := len(r.players)
		code := r.Code
		r.mu.Unlock()
		log.Printf("game: repko start rejected (room=%s reason=not_enough_players players=%d)", code, count)
		return ErrNotEnoughPlayers
	}
	board := NewBoard(newRNG())
	ps := make([]*PlayerState, 0, len(r.players))
	for _, p := range r.players {
		ps = append(ps, newPlayerState(p.ID, p.Name, p.Color))
	}
	civs := append([]Civilization(nil), Civilizations...)
	r.state = &GameState{
		Phase:               PhaseCivPick,
		Board:               board,
		Players:             ps,
		Current:             nil,
		Civilizations:       civs,
		PickedCivs:          map[string]string{},
		UsedActionsThisTurn: map[string]int{},
		PendingDiplomacy:    make([]DiplomacyOffer, 0),
	}
	code := r.Code
	tiles := len(board.Tiles)
	r.mu.Unlock()
	log.Printf("game: repko started (room=%s players=%d tiles=%d civilizations=%d)", code, len(ps), tiles, len(civs))
	go r.Broadcast()
	return nil
}

// Broadcast renders the State per-player (private "you" field) and fans out.
// Non-blocking send: drops on full channels.
func (r *Room) Broadcast() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	recipients := make([]*Player, len(r.players))
	copy(recipients, r.players)
	snapshots := r.buildSnapshotsLocked()
	r.mu.Unlock()

	for _, p := range recipients {
		b, ok := snapshots[p.ID]
		if !ok {
			continue
		}
		select {
		case p.SendCh <- b:
		default:
		}
	}
}

func (r *Room) buildSnapshotsLocked() map[string][]byte {
	out := make(map[string][]byte, len(r.players))
	if r.state == nil {
		players := make([]*PlayerState, 0, len(r.players))
		for _, p := range r.players {
			players = append(players, newPlayerState(p.ID, p.Name, p.Color))
		}
		base := State{
			Type:    MsgState,
			Phase:   string(PhaseLobby),
			Players: players,
		}
		for _, p := range r.players {
			env := base
			b, _ := json.Marshal(env)
			out[p.ID] = b
		}
		return out
	}

	tileCounts := r.state.tileCountByPlayer()
	unitCounts := make(map[string]int, len(r.state.Players))
	for _, t := range r.state.Board.Tiles {
		if t.OwnerID == "" {
			continue
		}
		unitCounts[t.OwnerID] += len(t.Garrison)
	}
	for _, ps := range r.state.Players {
		ps.TileCount = tileCounts[ps.ID]
		ps.UnitCount = unitCounts[ps.ID]
	}

	pending := r.state.PendingDiplomacy
	if pending == nil {
		pending = make([]DiplomacyOffer, 0)
	}

	base := State{
		Type:             MsgState,
		Phase:            string(r.state.Phase),
		Board:            r.state.Board,
		Players:          r.state.Players,
		Current:          r.state.Current,
		PendingDiplomacy: pending,
		WinnerID:         r.state.WinnerID,
	}
	if r.state.Phase == PhaseCivPick {
		base.Civilizations = r.state.Civilizations
	}
	for _, p := range r.players {
		env := base
		ps := r.state.playerByID(p.ID)
		if ps != nil {
			res := make(ResourceBank, len(ps.Resources))
			for k, v := range ps.Resources {
				res[k] = v
			}
			env.You = &YouState{Resources: res}
		}
		b, _ := json.Marshal(env)
		out[p.ID] = b
	}
	return out
}

// Close closes all player SendChs and marks the room closed. Idempotent.
func (r *Room) Close() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	players := make([]*Player, len(r.players))
	copy(players, r.players)
	r.mu.Unlock()

	errMsg, _ := json.Marshal(ErrorMsg{Type: MsgError, Message: "room closed"})
	for _, p := range players {
		select {
		case p.SendCh <- errMsg:
		default:
		}
		p.CloseSend()
	}
}

func (r *Room) emptySince() *time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.LastEmptyAt
}
