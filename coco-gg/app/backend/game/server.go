package game

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"sync"
)

var ErrUnknownRoom = errors.New("unknown room")
var ErrCodeExhausted = errors.New("could not allocate room code")

const maxCodeAttempts = 100

// Manager owns all active rooms and brokers join/create.
type Manager struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

// JoinOrCreate atomically:
//   - if code == "" : creates a fresh room with a unique generated code.
//   - if code != "" : joins the existing room with that code (or returns
//     ErrUnknownRoom if no such room exists).
func (m *Manager) JoinOrCreate(code, name string) (*Room, *Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var room *Room
	if code == "" {
		for i := 0; i < maxCodeAttempts; i++ {
			c := NewCode()
			if _, exists := m.rooms[c]; !exists {
				room = NewRoom(c)
				m.rooms[c] = room
				break
			}
		}
		if room == nil {
			return nil, nil, ErrCodeExhausted
		}
	} else {
		r, ok := m.rooms[code]
		if !ok {
			return nil, nil, ErrUnknownRoom
		}
		room = r
	}

	player := newPlayer(name, room.Arena)
	if !room.Add(player) {
		return nil, nil, ErrUnknownRoom
	}
	return room, player, nil
}

// Leave removes the player from the named room and deletes the room from the
// manager if it became empty.
func (m *Manager) Leave(roomCode, playerID string) {
	m.mu.Lock()
	room, ok := m.rooms[roomCode]
	m.mu.Unlock()
	if !ok {
		return
	}
	if room.Remove(playerID) {
		m.mu.Lock()
		// Only delete if it's still the same room instance we removed from.
		if cur, ok := m.rooms[roomCode]; ok && cur == room {
			delete(m.rooms, roomCode)
		}
		m.mu.Unlock()
	}
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
