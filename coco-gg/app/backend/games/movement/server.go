package movement

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"sync"
	"time"
)

var ErrUnknownRoom = errors.New("unknown room")
var ErrUnknownPlayer = errors.New("unknown player")
var ErrCodeExhausted = errors.New("could not allocate room code")

const (
	maxCodeAttempts = 100
	sweepInterval   = 30 * time.Second
	emptyRoomTTL    = 10 * time.Minute
)

// RoomDigest is the externally-visible per-room state for the dashboard.
type RoomDigest struct {
	Code      string         `json:"code"`
	Players   []PlayerDigest `json:"players"`
	CreatedAt int64          `json:"createdAt"`
}

type PlayerDigest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Stats struct {
	ActiveRooms  int `json:"activeRooms"`
	TotalPlayers int `json:"totalPlayers"`
}

type ManagerSnapshot struct {
	Rooms []RoomDigest `json:"rooms"`
	Stats Stats        `json:"stats"`
}

// Manager owns all active rooms.
type Manager struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

// CreateRoom allocates a unique 6-char code and registers a new room.
func (m *Manager) CreateRoom() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < maxCodeAttempts; i++ {
		c := NewCode()
		if _, exists := m.rooms[c]; !exists {
			m.rooms[c] = NewRoom(c)
			log.Printf("game: room created (code=%s)", c)
			return c, nil
		}
	}
	return "", ErrCodeExhausted
}

// Join attaches a new player to an existing room.
func (m *Manager) Join(code, name string) (*Room, *Player, error) {
	if code == "" {
		return nil, nil, ErrUnknownRoom
	}
	m.mu.Lock()
	room, ok := m.rooms[code]
	m.mu.Unlock()
	if !ok {
		return nil, nil, ErrUnknownRoom
	}
	player := newPlayer(name, room.Arena)
	if !room.Add(player) {
		return nil, nil, ErrUnknownRoom
	}
	log.Printf("game: player joined (room=%s player_id=%s name=%q)", room.Code, player.ID, player.Name)
	return room, player, nil
}

// GetRoom returns the room with the given code, if it exists.
func (m *Manager) GetRoom(code string) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[code]
	return r, ok
}

// DestroyRoom removes the room from the manager and closes it. Idempotent.
// Does not hold the manager mutex while Close runs (Close acquires the room
// mutex and writes to player channels — different lock domain).
func (m *Manager) DestroyRoom(code string) {
	m.mu.Lock()
	room, ok := m.rooms[code]
	if ok {
		delete(m.rooms, code)
	}
	m.mu.Unlock()
	if !ok {
		return
	}
	log.Printf("game: room destroyed (code=%s reason=explicit)", code)
	room.Close()
}

// Kick removes the named player from the named room with the given reason.
// Returns ErrUnknownRoom if the code is invalid, ErrUnknownPlayer if the
// player is not in that room. The room continues to exist even when empty;
// the sweeper will collect forgotten empties.
func (m *Manager) Kick(roomCode, playerID, reason string) error {
	m.mu.Lock()
	room, ok := m.rooms[roomCode]
	m.mu.Unlock()
	if !ok {
		return ErrUnknownRoom
	}
	if !room.Kick(playerID, reason) {
		return ErrUnknownPlayer
	}
	log.Printf("game: player kicked (room=%s player_id=%s reason=%q)", roomCode, playerID, reason)
	return nil
}

// Leave removes the player from the named room without deleting the room.
func (m *Manager) Leave(roomCode, playerID string) {
	m.mu.Lock()
	room, ok := m.rooms[roomCode]
	m.mu.Unlock()
	if !ok {
		return
	}
	room.Remove(playerID)
	log.Printf("game: player left (room=%s player_id=%s)", roomCode, playerID)
}

// Snapshot returns the dashboard view of every active room. Nested locks:
// manager mu → room mu, never the reverse.
func (m *Manager) Snapshot() ManagerSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	rooms := make([]RoomDigest, 0, len(m.rooms))
	totalPlayers := 0
	for _, r := range m.rooms {
		d := r.Digest()
		totalPlayers += len(d.Players)
		rooms = append(rooms, d)
	}
	return ManagerSnapshot{
		Rooms: rooms,
		Stats: Stats{ActiveRooms: len(rooms), TotalPlayers: totalPlayers},
	}
}

// StartSweeper GCs rooms that have been empty for longer than emptyRoomTTL.
// Same nested-lock discipline as Snapshot.
func (m *Manager) StartSweeper(ctx context.Context) {
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.sweep()
		}
	}
}

func (m *Manager) sweep() {
	for _, e := range m.collectExpired() {
		log.Printf("game: sweeper deleted room (code=%s empty_for=%s)", e.code, time.Since(*e.emptyAt))
		m.removeAndClose(e.code)
	}
}

type expiredRoom struct {
	code    string
	emptyAt *time.Time
}

func (m *Manager) collectExpired() []expiredRoom {
	m.mu.Lock()
	defer m.mu.Unlock()
	var expired []expiredRoom
	for code, r := range m.rooms {
		if t := r.emptySince(); t != nil && time.Since(*t) > emptyRoomTTL {
			expired = append(expired, expiredRoom{code: code, emptyAt: t})
		}
	}
	return expired
}

func (m *Manager) removeAndClose(code string) {
	m.mu.Lock()
	room, ok := m.rooms[code]
	if ok {
		delete(m.rooms, code)
	}
	m.mu.Unlock()
	if !ok {
		return
	}
	room.Close()
}

func newPlayer(name string, arena Bounds) *Player {
	id := newPlayerID()
	cx := arena.W / 2
	cy := arena.H / 2
	ox, oy := randomOffset(50.0)
	return &Player{
		ID:     id,
		Name:   name,
		X:      cx + ox,
		Y:      cy + oy,
		Color:  colorFromID(id),
		SendCh: make(chan []byte, sendChCapacity),
	}
}

func newPlayerID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

func randomOffset(span float64) (float64, float64) {
	return randUniform(span), randUniform(span)
}

func randUniform(span float64) float64 {
	n, err := rand.Int(rand.Reader, big.NewInt(2001))
	if err != nil {
		panic(err)
	}
	return (float64(n.Int64())/1000.0 - 1.0) * span
}

// colorFromID maps the first byte of the hex player ID to an HSL hue (0-360),
// then converts HSL(hue, 70%, 55%) to a stable #RRGGBB string.
func colorFromID(id string) string {
	if len(id) < 2 {
		return "#888888"
	}
	var b byte
	fmt.Sscanf(id[:2], "%02x", &b)
	hue := float64(b) / 255.0 * 360.0
	r, g, bl := hslToRGB(hue, 0.70, 0.55)
	return fmt.Sprintf("#%02X%02X%02X", r, g, bl)
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	c := (1 - math.Abs(2*l-1)) * s
	hp := h / 60.0
	x := c * (1 - math.Abs(math.Mod(hp, 2)-1))
	var r1, g1, b1 float64
	switch {
	case hp < 1:
		r1, g1, b1 = c, x, 0
	case hp < 2:
		r1, g1, b1 = x, c, 0
	case hp < 3:
		r1, g1, b1 = 0, c, x
	case hp < 4:
		r1, g1, b1 = 0, x, c
	case hp < 5:
		r1, g1, b1 = x, 0, c
	default:
		r1, g1, b1 = c, 0, x
	}
	m := l - c/2
	return uint8(math.Round((r1 + m) * 255)),
		uint8(math.Round((g1 + m) * 255)),
		uint8(math.Round((b1 + m) * 255))
}
