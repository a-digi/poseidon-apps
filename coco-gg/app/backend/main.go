package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"coco-gg-plugin/game"
)

const (
	pluginID    = "coco-gg"
	portFile    = "coco-gg.port"
	portFileTmp = "coco-gg.port.tmp"
)

func main() {
	dataDir := resolveDataDir()
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("[coco-gg] mkdir data dir: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("[coco-gg] listen: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	if err := writePortFile(dataDir, port); err != nil {
		listener.Close()
		log.Fatalf("[coco-gg] write port file: %v", err)
	}
	defer os.Remove(filepath.Join(dataDir, portFile))

	log.Printf("[coco-gg] listening on 127.0.0.1:%d", port)

	mgr := game.NewManager()
	wsh := NewWSHandler(mgr)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", wsh.ServeHTTP)

	server := &http.Server{
		Handler:     mux,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	serveErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case <-ctx.Done():
	case err := <-serveErr:
		if err != nil {
			log.Printf("[coco-gg] serve error: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true,"pluginId":"coco-gg"}`))
}

func resolveDataDir() string {
	if dir := os.Getenv("PLUGIN_DATA_DIR"); dir != "" {
		return dir
	}
	return "./data"
}

func writePortFile(dataDir string, port int) error {
	tmp := filepath.Join(dataDir, portFileTmp)
	if err := os.WriteFile(tmp, []byte(strconv.Itoa(port)), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dataDir, portFile))
}
