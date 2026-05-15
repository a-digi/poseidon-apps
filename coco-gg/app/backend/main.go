package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"coco-gg-plugin/deploy/local"
	"coco-gg-plugin/deploy/remote"
	"coco-gg-plugin/games/movement"
	"coco-gg-plugin/games/repko"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var err error
	switch os.Getenv("MODE") {
	case "remote":
		err = remote.Run(ctx, []remote.GameRegistrar{movement.Register, repko.Register})
	default:
		err = local.Run(ctx, []local.GameRegistrar{movement.Register, repko.Register})
	}
	if err != nil {
		log.Fatalf("run: %v", err)
	}
}
