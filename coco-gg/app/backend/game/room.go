package game

import (
	"encoding/json"
	"sync"
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
}

// Room owns its player set, position state, and tick goroutine.
type Room struct {
	Code    string
	Arena   Bounds
	mu      sync.Mutex
	players map[string]*Player
	stopCh  chan struct{}
	stopped bool
}

// NewRoom constructs a room and starts its tick goroutine.
func NewRoom(code string) *Room {
	r := &Room{
		Code:    code,
		Arena:   Bounds{W: 800, H: 600},
		players: make(map[string]*Player),
		stopCh:  make(chan struct{}),
	}
	go r.run()
	return r
}

// Add inserts a player. Returns false if the room is already being torn down.
func (r *Room) Add(p *Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stopped {
		return false
	}
	r.players[p.ID] = p
	return true
}

// Remove pulls a player out and broadcasts {type: "left"}. Returns true if the
// room is now empty (caller — the Manager — should delete the room).
func (r *Room) Remove(playerID string) (empty bool) {
	r.mu.Lock()
	if _, ok := r.players[playerID]; !ok {
		empty = len(r.players) == 0
		r.mu.Unlock()
		return empty
	}
	delete(r.players, playerID)

	leftMsg, _ := json.Marshal(Left{Type: MsgLeft, PlayerID: playerID})
	for _, p := range r.players {
		select {
		case p.SendCh <- leftMsg:
		default:
		}
	}
	empty = len(r.players) == 0
	if empty {
		r.stopped = true
		close(r.stopCh)
	}
	r.mu.Unlock()
	return empty
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
