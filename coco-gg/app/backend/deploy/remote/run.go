package remote

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"coco-gg-plugin/backend/gamemeta"
	"coco-gg-plugin/backend/runtime"
)

type GameRegistrar func(ctx context.Context, r runtime.Router) gamemeta.Info

func Run(ctx context.Context, registrars []GameRegistrar) error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		return errors.New("remote mode requires ADMIN_TOKEN env")
	}
	publicURL := os.Getenv("PUBLIC_URL")
	if publicURL == "" {
		return errors.New("remote mode requires PUBLIC_URL env")
	}
	uiDir := os.Getenv("UI_DIR")
	if uiDir == "" {
		uiDir = "./ui"
	}

	log.SetPrefix("[coco-gg] ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("main: starting (mode=remote port=%s public_url=%s ui_dir=%s)", port, publicURL, uiDir)

	sessions := NewSessionStore()
	go sessions.StartSweeper(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	adminMW := requireAdminToken(adminToken)
	publicMW := requireMobileSession(sessions)

	r := &router{mux: mux, wrapAdmin: adminMW, wrapPublic: publicMW}
	games := make([]gamemeta.Info, 0, len(registrars))
	for _, reg := range registrars {
		games = append(games, reg(ctx, r))
	}

	mux.Handle("GET /api/games", adminMW(http.HandlerFunc(listGamesHandler(games))))
	mux.Handle("POST /api/admin/sessions", adminMW(http.HandlerFunc(MintSessionHandler(sessions))))

	mux.Handle("/admin/", adminMW(uiHandler(uiDir, "/admin/")))
	mux.Handle("/plugins/coco-gg/", uiHandler(uiDir, "/plugins/coco-gg/"))

	server := &http.Server{
		Addr:        ":" + port,
		Handler:     mux,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}
	serveErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()
	log.Printf("main: listening (addr=:%s)", port)

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
