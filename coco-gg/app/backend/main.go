package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"coco-gg-plugin/backend/deploy/local"
	"coco-gg-plugin/backend/deploy/remote"
	"coco-gg-plugin/games/repko/backend"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var err error
	switch os.Getenv("MODE") {
	case "remote":
		err = remote.Run(ctx, []remote.GameRegistrar{repko.Register})
	default:
		err = local.Run(ctx, []local.GameRegistrar{repko.Register})
	}
	if err != nil {
		log.Fatalf("run: %v", err)
	}
}
