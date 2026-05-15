package movement

import (
	"context"

	"coco-gg-plugin/gamemeta"
	"coco-gg-plugin/runtime"
)

// Register mounts the movement game's HTTP and WebSocket handlers via the
// role-aware runtime.Router, starts its sweeper goroutine, and returns the
// metadata to be advertised via GET /api/games.
func Register(ctx context.Context, r runtime.Router) gamemeta.Info {
	mgr := NewManager()
	go mgr.StartSweeper(ctx)

	wsh := newWSHandler(mgr)
	r.Public("GET /ws/games/movement", wsh)

	r.Admin("POST /api/games/movement/rooms", createRoomHandler(mgr))
	r.Admin("GET /api/games/movement/rooms", listRoomsHandler(mgr))
	r.Admin("GET /api/games/movement/rooms/{code}", getRoomHandler(mgr))
	r.Admin("DELETE /api/games/movement/rooms/{code}", destroyRoomHandler(mgr))
	r.Admin("DELETE /api/games/movement/rooms/{code}/players/{playerId}", kickPlayerHandler(mgr))

	return gamemeta.Info{
		ID:          "movement",
		Name:        "Movement Arena",
		Description: "A shared 2D arena where players move colored circles. Placeholder scaffold for richer mechanics.",
	}
}
