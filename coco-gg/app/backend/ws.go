package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"coco-gg-plugin/game"

	"github.com/coder/websocket"
)

type WSHandler struct {
	mgr *game.Manager
}

func NewWSHandler(mgr *game.Manager) *WSHandler {
	return &WSHandler{mgr: mgr}
}

func (h *WSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// InsecureSkipVerify: the host's reverse proxy strips the Origin header
	// before forwarding, so origin checks here would always fail. Auth is
	// enforced at the host edge by the mobile-token middleware.
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer c.Close(websocket.StatusInternalError, "")

	ctx := r.Context()

	_, raw, err := c.Read(ctx)
	if err != nil {
		return
	}
	var hello game.Hello
	if json.Unmarshal(raw, &hello) != nil || hello.Type != game.MsgHello {
		writeError(ctx, c, "expected hello")
		return
	}
	name := strings.TrimSpace(hello.Name)
	if len(name) < 1 || len(name) > 32 {
		writeError(ctx, c, "name must be 1-32 characters")
		return
	}

	room, player, err := h.mgr.JoinOrCreate(hello.Room, name)
	if err != nil {
		writeError(ctx, c, err.Error())
		return
	}

	welcome := game.Welcome{
		Type:     game.MsgWelcome,
		PlayerID: player.ID,
		Room:     room.Code,
		TickHz:   20,
		Arena:    room.Arena,
	}
	welcomeBytes, _ := json.Marshal(welcome)
	if err := c.Write(ctx, websocket.MessageText, welcomeBytes); err != nil {
		h.mgr.Leave(room.Code, player.ID)
		return
	}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-player.SendCh:
				if !ok {
					return
				}
				if err := c.Write(ctx, websocket.MessageText, msg); err != nil {
					return
				}
			}
		}
	}()

	for {
		_, raw, err := c.Read(ctx)
		if err != nil {
			break
		}
		var in game.Input
		if json.Unmarshal(raw, &in) != nil || in.Type != game.MsgInput {
			continue
		}
		room.SetInput(player.ID, in)
	}

	h.mgr.Leave(room.Code, player.ID)
	close(player.SendCh)
	<-writerDone
	c.Close(websocket.StatusNormalClosure, "")
}

func writeError(ctx context.Context, c *websocket.Conn, msg string) {
	b, _ := json.Marshal(game.ErrorMsg{Type: game.MsgError, Message: msg})
	_ = c.Write(ctx, websocket.MessageText, b)
}
