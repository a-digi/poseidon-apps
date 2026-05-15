package local

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"coco-gg-plugin/backend/gamemeta"
	"coco-gg-plugin/backend/runtime"
)

const (
	portFile    = "coco-gg.port"
	portFileTmp = "coco-gg.port.tmp"
)

// GameRegistrar is the signature a game's Register function exposes.
type GameRegistrar func(ctx context.Context, r runtime.Router) gamemeta.Info

// Run is the local-mode entry point: spawned by the Wails host's
// longrunning registry, binds 127.0.0.1, writes the port file to
// ${PLUGIN_DATA_DIR}/coco-gg.port, runs until ctx is done.
func Run(ctx context.Context, registrars []GameRegistrar) error {
	dataDir := resolveDataDir()
	closer, err := setupLogging(dataDir)
	if err != nil {
		return err
	}
	defer closer.Close()

	log.Printf("main: starting (mode=local data_dir=%s)", dataDir)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := writePortFile(dataDir, port); err != nil {
		listener.Close()
		return err
	}
	defer os.Remove(filepath.Join(dataDir, portFile))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	r := &router{mux: mux}
	games := make([]gamemeta.Info, 0, len(registrars))
	for _, reg := range registrars {
		games = append(games, reg(ctx, r))
	}
	mux.HandleFunc("GET /api/games", listGamesHandler(games))

	server := &http.Server{
		Handler:     mux,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	serveErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()
	log.Printf("main: listening (addr=%s)", listener.Addr())

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	log.Printf("main: signal received, shutting down")
	sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(sctx); err != nil {
		return err
	}
	log.Printf("main: shutdown complete")
	return nil
}

func resolveDataDir() string {
	if d := os.Getenv("PLUGIN_DATA_DIR"); d != "" {
		return d
	}
	return "./data"
}

func writePortFile(dataDir string, port int) error {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	tmp := filepath.Join(dataDir, portFileTmp)
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(port)), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dataDir, portFile))
}

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

