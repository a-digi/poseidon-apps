package movement

import (
	"context"
	"net/http"

	"coco-gg-plugin/gamemeta"
)

// Register mounts the movement game's HTTP and WebSocket handlers on the
// shared mux, starts its sweeper goroutine, and returns the metadata to be
// advertised via GET /api/games.
func Register(ctx context.Context, mux *http.ServeMux) gamemeta.Info {
	mgr := NewManager()
	go mgr.StartSweeper(ctx)

	wsh := newWSHandler(mgr)
	mux.HandleFunc("GET /ws/games/movement", wsh.ServeHTTP)

	mux.HandleFunc("POST /api/games/movement/rooms", createRoomHandler(mgr))
	mux.HandleFunc("GET /api/games/movement/rooms", listRoomsHandler(mgr))
	mux.HandleFunc("GET /api/games/movement/rooms/{code}", getRoomHandler(mgr))
	mux.HandleFunc("DELETE /api/games/movement/rooms/{code}", destroyRoomHandler(mgr))
	mux.HandleFunc("DELETE /api/games/movement/rooms/{code}/players/{playerId}", kickPlayerHandler(mgr))

	return gamemeta.Info{
		ID:          "movement",
		Name:        "Movement Arena",
		Description: "A shared 2D arena where players move colored circles. Placeholder scaffold for richer mechanics.",
	}
}
