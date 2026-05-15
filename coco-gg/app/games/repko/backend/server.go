package repko

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var (
	ErrUnknownRoom      = errors.New("unknown room")
	ErrUnknownPlayer    = errors.New("unknown player")
	ErrCodeExhausted    = errors.New("could not generate unique room code")
	ErrAlreadyStarted   = errors.New("game already started")
	ErrNotEnoughPlayers = errors.New("need at least 2 players to start")
)

const (
	maxCodeAttempts = 100
	sweepInterval   = 30 * time.Second
	emptyRoomTTL    = 10 * time.Minute
)

type RoomDigest struct {
	Code      string         `json:"code"`
	Phase     string         `json:"phase"`
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
	ActiveGames  int `json:"activeGames"`
}

type ManagerSnapshot struct {
	Rooms []RoomDigest `json:"rooms"`
	Stats Stats        `json:"stats"`
}

type Manager struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{rooms: make(map[string]*Room)}
}

func (m *Manager) CreateRoom(expectedPlayers, maxRounds int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := 0; i < maxCodeAttempts; i++ {
		c := NewCode()
		if _, exists := m.rooms[c]; !exists {
			m.rooms[c] = NewRoom(c, expectedPlayers, maxRounds)
			log.Printf("game: room created (code=%s expectedPlayers=%d maxRounds=%d)", c, expectedPlayers, maxRounds)
			return c, nil
		}
	}
	return "", ErrCodeExhausted
}

func (m *Manager) GetRoom(code string) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rooms[code]
	return r, ok
}

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

// Snapshot returns the dashboard view of every active room. Nested locks:
// manager mu -> room mu, never the reverse.
func (m *Manager) Snapshot() ManagerSnapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	rooms := make([]RoomDigest, 0, len(m.rooms))
	totalPlayers := 0
	activeGames := 0
	for _, r := range m.rooms {
		d := r.Digest()
		totalPlayers += len(d.Players)
		if d.Phase != "lobby" {
			activeGames++
		}
		rooms = append(rooms, d)
	}
	return ManagerSnapshot{
		Rooms: rooms,
		Stats: Stats{
			ActiveRooms:  len(rooms),
			TotalPlayers: totalPlayers,
			ActiveGames:  activeGames,
		},
	}
}

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

// StartGame transitions a room from lobby to setup1.
func (m *Manager) StartGame(code string) error {
	m.mu.Lock()
	room, ok := m.rooms[code]
	m.mu.Unlock()
	if !ok {
		return ErrUnknownRoom
	}
	return room.StartGame()
}

// Kick removes a player from a room with the given reason. Returns
// ErrUnknownRoom or ErrUnknownPlayer if the lookup fails.
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
