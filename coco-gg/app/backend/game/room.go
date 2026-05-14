package game

import (
	"encoding/json"
	"sync"
	"time"
)

const sendChCapacity = 16

// Player holds the runtime state for one connected client.
// All fields are guarded by the owning Room's mutex except SendCh (channel ops are safe).
type Player struct {
	ID        string
	Name      string
	X         float64
	Y         float64
	Color     string
	LastInput Input
	SendCh    chan []byte
	closeOnce sync.Once
}

// CloseSend closes the player's send channel exactly once. Safe to call from
// both the connection handler (normal disconnect) and Room.Close (forced
// shutdown) without panicking on double-close.
func (p *Player) CloseSend() {
	p.closeOnce.Do(func() { close(p.SendCh) })
}

// Room owns its player set, position state, and tick goroutine.
type Room struct {
	Code        string
	Arena       Bounds
	CreatedAt   time.Time
	LastEmptyAt *time.Time
	mu          sync.Mutex
	players     map[string]*Player
	stopCh      chan struct{}
	closed      bool
}

// NewRoom constructs a room and starts its tick goroutine.
func NewRoom(code string) *Room {
	now := time.Now()
	r := &Room{
		Code:        code,
		Arena:       Bounds{W: 800, H: 600},
		CreatedAt:   now,
		LastEmptyAt: &now,
		players:     make(map[string]*Player),
		stopCh:      make(chan struct{}),
	}
	go r.run()
	return r
}

// Add inserts a player. Returns false if the room is already being torn down.
func (r *Room) Add(p *Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return false
	}
	wasEmpty := len(r.players) == 0
	r.players[p.ID] = p
	if wasEmpty {
		r.LastEmptyAt = nil
	}
	return true
}

// Remove pulls a player out and broadcasts {type: "left"} to remaining players.
// Marks the room as empty when the last player leaves; never auto-deletes.
func (r *Room) Remove(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.players[playerID]; !ok {
		return
	}
	delete(r.players, playerID)
	if r.closed {
		return
	}

	leftMsg, _ := json.Marshal(Left{Type: MsgLeft, PlayerID: playerID})
	for _, p := range r.players {
		select {
		case p.SendCh <- leftMsg:
		default:
		}
	}
	if len(r.players) == 0 {
		now := time.Now()
		r.LastEmptyAt = &now
	}
}

// Kick removes a player by ID. Sends them an {type:"error", message:reason}
// frame immediately before closing their send channel; peers receive the
// normal {type:"left"} broadcast via the existing Remove path. Returns true
// if a player was actually kicked; false if the room is closed or the
// player isn't present.
//
// Lock discipline: holds the room mutex briefly to enqueue the error frame
// and to read the player ref, then releases it before calling CloseSend
// (which is idempotent via sync.Once) and Remove (which re-acquires the
// mutex). Matches the same nesting Close() uses.
func (r *Room) Kick(playerID, reason string) bool {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return false
	}
	p, ok := r.players[playerID]
	if !ok {
		r.mu.Unlock()
		return false
	}
	errMsg, _ := json.Marshal(ErrorMsg{Type: MsgError, Message: reason})
	select {
	case p.SendCh <- errMsg:
	default:
	}
	r.mu.Unlock()

	p.CloseSend()
	r.Remove(playerID)
	return true
}

// Close terminates the room: notifies any remaining players, closes their send
// channels, and stops the tick goroutine. Safe to call more than once; the
// closed guard protects against DestroyRoom racing the sweeper.
func (r *Room) Close() {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return
	}
	r.closed = true
	players := make([]*Player, 0, len(r.players))
	for _, p := range r.players {
		players = append(players, p)
	}
	r.mu.Unlock()

	errMsg, _ := json.Marshal(ErrorMsg{Type: MsgError, Message: "room closed"})
	for _, p := range players {
		select {
		case p.SendCh <- errMsg:
		default:
		}
		p.CloseSend()
	}
	close(r.stopCh)
}

// SetInput stores the player's latest input under the room mutex.
func (r *Room) SetInput(playerID string, in Input) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.players[playerID]; ok {
		p.LastInput = in
	}
}

// Snapshot builds a Snapshot value from the current state. Caller holds no
// locks; the function takes and releases the room mutex.
func (r *Room) Snapshot(tick int) Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.snapshotLocked(tick)
}

func (r *Room) snapshotLocked(tick int) Snapshot {
	players := make([]PlayerSnapshot, 0, len(r.players))
	for _, p := range r.players {
		players = append(players, PlayerSnapshot{
			ID:    p.ID,
			Name:  p.Name,
			X:     p.X,
			Y:     p.Y,
			Color: p.Color,
		})
	}
	return Snapshot{Type: MsgSnapshot, Tick: tick, Players: players}
}

// Digest returns the externally-visible state of the room (id+name only, no
// channels or connection objects). Acquires the room mutex.
func (r *Room) Digest() RoomDigest {
	r.mu.Lock()
	defer r.mu.Unlock()
	players := make([]PlayerDigest, 0, len(r.players))
	for _, p := range r.players {
		players = append(players, PlayerDigest{ID: p.ID, Name: p.Name})
	}
	return RoomDigest{
		Code:      r.Code,
		Players:   players,
		CreatedAt: r.CreatedAt.Unix(),
	}
}

// emptySince returns the value of LastEmptyAt under the room mutex.
func (r *Room) emptySince() *time.Time {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.LastEmptyAt
}
