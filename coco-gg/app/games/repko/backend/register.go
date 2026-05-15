package repko

import (
	"context"

	"coco-gg-plugin/backend/gamemeta"
	"coco-gg-plugin/backend/runtime"
)

func Register(ctx context.Context, r runtime.Router) gamemeta.Info {
	mgr := NewManager()
	go mgr.StartSweeper(ctx)

	wsh := newWSHandler(mgr)
	r.Public("GET /ws/games/repko", wsh)
	r.Public("POST /api/games/repko/rooms/{code}/leave", leaveRoomHandler(mgr))

	r.Admin("POST /api/games/repko/rooms", createRoomHandler(mgr))
	r.Admin("GET /api/games/repko/rooms", listRoomsHandler(mgr))
	r.Admin("GET /api/games/repko/rooms/{code}", getRoomHandler(mgr))
	r.Admin("DELETE /api/games/repko/rooms/{code}", destroyRoomHandler(mgr))
	r.Admin("DELETE /api/games/repko/rooms/{code}/players/{playerId}", kickPlayerHandler(mgr))
	r.Admin("POST /api/games/repko/rooms/{code}/start", startGameHandler(mgr))

	return gamemeta.Info{
		ID:          "repko",
		Name:        "Repko",
		Description: "A turn-based hex-board game of resource gathering and settlement, for 3-6 players.",
	}
}
