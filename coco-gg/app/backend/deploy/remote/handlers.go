package remote

import (
	"encoding/json"
	"log"
	"net/http"

	"coco-gg-plugin/gamemeta"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("http: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "pluginId": "coco-gg"})
	log.Printf("http: %s %s -> 200", r.Method, r.URL.Path)
}

func listGamesHandler(games []gamemeta.Info) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("http: %s %s", r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(games); err != nil {
			log.Printf("http: %s %s -> 500 (err=%v)", r.Method, r.URL.Path, err)
			return
		}
		log.Printf("http: %s %s -> 200 (games=%d)", r.Method, r.URL.Path, len(games))
	}
}

