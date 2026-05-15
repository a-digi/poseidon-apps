package movement

import (
	"encoding/json"
	"math"
	"time"
)

const (
	tickHz       = 20
	playerSpeed  = 200.0
	tickInterval = time.Second / tickHz
)

// run is the per-room tick goroutine entry point. Started by NewRoom().
// Exits when r.stopCh is closed.
func (r *Room) run() {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	const dt = 1.0 / float64(tickHz)
	tick := 0
	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			tick++
			snap := r.step(tick, dt)
			bytes, _ := json.Marshal(snap)
			r.fanout(bytes)
		}
	}
}

func (r *Room) step(tick int, dt float64) Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.players {
		dx, dy := 0.0, 0.0
		if p.LastInput.Right {
			dx += 1
		}
		if p.LastInput.Left {
			dx -= 1
		}
		if p.LastInput.Down {
			dy += 1
		}
		if p.LastInput.Up {
			dy -= 1
		}
		p.X = math.Min(math.Max(p.X+dx*playerSpeed*dt, 0), r.Arena.W)
		p.Y = math.Min(math.Max(p.Y+dy*playerSpeed*dt, 0), r.Arena.H)
	}
	return r.snapshotLocked(tick)
}

func (r *Room) fanout(msg []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	for _, p := range r.players {
		select {
		case p.SendCh <- msg:
		default:
		}
	}
}
