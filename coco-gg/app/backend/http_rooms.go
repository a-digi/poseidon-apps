package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"

	"coco-gg-plugin/game"
)

var validCode = regexp.MustCompile(`^[A-HJ-NP-Z2-9]{6}$`)
var validPlayerID = regexp.MustCompile(`^[0-9a-f]{8}$`)

func CreateRoomHandler(mgr *game.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		code, err := mgr.CreateRoom()
		if errors.Is(err, game.ErrCodeExhausted) {
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

func ListRoomsHandler(mgr *game.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		log.Printf("http: %s %s -> %d (code=%s)", r.Method, r.URL.Path, http.StatusOK, "")
		_ = json.NewEncoder(w).Encode(mgr.Snapshot())
	}
}

func GetRoomHandler(mgr *game.Manager) http.HandlerFunc {
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

func KickPlayerHandler(mgr *game.Manager) http.HandlerFunc {
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
		if errors.Is(err, game.ErrUnknownRoom) {
			log.Printf("http: %s %s -> %d (err=unknown_room)", r.Method, r.URL.Path, http.StatusNotFound)
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unknown room"})
			return
		}
		log.Printf("http: %s %s -> %d (code=%s)", r.Method, r.URL.Path, http.StatusNoContent, code)
		w.WriteHeader(http.StatusNoContent)
	}
}

func DestroyRoomHandler(mgr *game.Manager) http.HandlerFunc {
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
