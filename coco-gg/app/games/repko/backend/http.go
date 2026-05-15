package repko

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
)

var validCode = regexp.MustCompile(`^[A-HJ-NP-Z2-9]{6}$`)
var validPlayerID = regexp.MustCompile(`^[0-9a-f]{8}$`)

func createRoomHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		var body struct {
			ExpectedPlayers int `json:"expectedPlayers"`
			MaxRounds       int `json:"maxRounds"`
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		expected := body.ExpectedPlayers
		if expected < 2 || expected > maxPlayers {
			expected = 0
		}
		maxRounds := body.MaxRounds
		if maxRounds < 1 || maxRounds > 100 {
			maxRounds = 0
		}
		code, err := mgr.CreateRoom(expected, maxRounds)
		if errors.Is(err, ErrCodeExhausted) {
			log.Printf("http: %s %s -> %d (err=%s)", r.Method, r.URL.Path, http.StatusServiceUnavailable, "no codes available")
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "no codes available"})
			return
		}
		log.Printf("http: %s %s -> %d (code=%s)", r.Method, r.URL.Path, http.StatusCreated, code)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"code": code})
	}
}

func listRoomsHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		log.Printf("http: %s %s -> %d (code=%s)", r.Method, r.URL.Path, http.StatusOK, "")
		_ = json.NewEncoder(w).Encode(mgr.Snapshot())
	}
}

func getRoomHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		code := r.PathValue("code")
		if !validCode.MatchString(code) {
			log.Printf("http: %s %s -> %d (err=%s)", r.Method, r.URL.Path, http.StatusBadRequest, "invalid code")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid code"})
			return
		}
		room, ok := mgr.GetRoom(code)
		if !ok {
			log.Printf("http: %s %s -> %d (err=%s)", r.Method, r.URL.Path, http.StatusNotFound, "not found")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
			return
		}
		log.Printf("http: %s %s -> %d (code=%s)", r.Method, r.URL.Path, http.StatusOK, code)
		_ = json.NewEncoder(w).Encode(room.Digest())
	}
}

func destroyRoomHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		code := r.PathValue("code")
		if !validCode.MatchString(code) {
			log.Printf("http: %s %s -> %d (err=%s)", r.Method, r.URL.Path, http.StatusBadRequest, "invalid code")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid code"})
			return
		}
		mgr.DestroyRoom(code)
		log.Printf("http: %s %s -> %d (code=%s)", r.Method, r.URL.Path, http.StatusNoContent, code)
		w.WriteHeader(http.StatusNoContent)
	}
}

func kickPlayerHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		code := r.PathValue("code")
		if !validCode.MatchString(code) {
			log.Printf("http: %s %s -> %d (err=invalid_code)", r.Method, r.URL.Path, http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid code"})
			return
		}
		playerID := r.PathValue("playerId")
		if !validPlayerID.MatchString(playerID) {
			log.Printf("http: %s %s -> %d (err=invalid_player_id)", r.Method, r.URL.Path, http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid player id"})
			return
		}
		err := mgr.Kick(code, playerID, "kicked")
		if errors.Is(err, ErrUnknownRoom) {
			log.Printf("http: %s %s -> %d (err=unknown_room)", r.Method, r.URL.Path, http.StatusNotFound)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unknown room"})
			return
		}
		log.Printf("http: %s %s -> %d (code=%s)", r.Method, r.URL.Path, http.StatusNoContent, code)
		w.WriteHeader(http.StatusNoContent)
	}
}

func leaveRoomHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		code := r.PathValue("code")
		if !validCode.MatchString(code) {
			log.Printf("http: %s %s -> %d (err=invalid_code)", r.Method, r.URL.Path, http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid code"})
			return
		}
		var body struct {
			ResumeToken string `json:"resumeToken"`
		}
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		if body.ResumeToken == "" {
			log.Printf("http: %s %s -> %d (err=missing_token)", r.Method, r.URL.Path, http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "resumeToken required"})
			return
		}
		room, ok := mgr.GetRoom(code)
		if !ok {
			log.Printf("http: %s %s -> %d (err=unknown_room)", r.Method, r.URL.Path, http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unknown room"})
			return
		}
		playerID, found := room.playerIDByResumeToken(body.ResumeToken)
		if !found {
			log.Printf("http: %s %s -> %d (err=token_unknown)", r.Method, r.URL.Path, http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "token unknown"})
			return
		}
		room.LeaveGame(playerID)
		log.Printf("http: %s %s -> %d (code=%s player_id=%s)", r.Method, r.URL.Path, http.StatusNoContent, code, playerID)
		w.WriteHeader(http.StatusNoContent)
	}
}

func startGameHandler(mgr *Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		code := r.PathValue("code")
		if !validCode.MatchString(code) {
			log.Printf("http: %s %s -> %d (err=invalid_code)", r.Method, r.URL.Path, http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid code"})
			return
		}
		err := mgr.StartGame(code)
		switch {
		case err == nil:
			log.Printf("http: %s %s -> %d (code=%s)", r.Method, r.URL.Path, http.StatusNoContent, code)
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, ErrUnknownRoom):
			log.Printf("http: %s %s -> %d (err=unknown_room)", r.Method, r.URL.Path, http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unknown room"})
		default:
			log.Printf("http: %s %s -> %d (err=%v)", r.Method, r.URL.Path, http.StatusConflict, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
	}
}
