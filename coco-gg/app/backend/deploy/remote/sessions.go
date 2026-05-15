package remote

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
)

const defaultSessionTTL = time.Hour

type SessionStore struct {
	mu       sync.Mutex
	sessions map[string]int64
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: map[string]int64{}}
}

func (s *SessionStore) Mint(ttl time.Duration) (string, int64, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", 0, err
	}
	token := hex.EncodeToString(buf)
	expiresAt := time.Now().Add(ttl).Unix()
	s.mu.Lock()
	s.sessions[token] = expiresAt
	s.mu.Unlock()
	return token, expiresAt, nil
}

func (s *SessionStore) Validate(token string) bool {
	s.mu.Lock()
	expiresAt, ok := s.sessions[token]
	s.mu.Unlock()
	if !ok {
		return false
	}
	if time.Now().Unix() > expiresAt {
		s.mu.Lock()
		delete(s.sessions, token)
		s.mu.Unlock()
		return false
	}
	return true
}

func (s *SessionStore) StartSweeper(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			now := time.Now().Unix()
			s.mu.Lock()
			for k, v := range s.sessions {
				if now > v {
					delete(s.sessions, k)
				}
			}
			s.mu.Unlock()
		}
	}
}

type mintSessionResp struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expiresAt"`
}

type mintSessionReq struct {
	TTLSeconds int `json:"ttlSeconds"`
}

func MintSessionHandler(store *SessionStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		var req mintSessionReq
		_ = json.NewDecoder(r.Body).Decode(&req)
		ttl := defaultSessionTTL
		if req.TTLSeconds > 0 {
			ttl = time.Duration(req.TTLSeconds) * time.Second
		}
		token, expiresAt, err := store.Mint(ttl)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("http: %s %s -> 500 (err=%v)", r.Method, r.URL.Path, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mintSessionResp{Token: token, ExpiresAt: expiresAt})
		log.Printf("http: %s %s -> 200 (ttl=%s)", r.Method, r.URL.Path, ttl)
	}
}
