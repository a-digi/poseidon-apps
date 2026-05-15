package runtime

import "net/http"

// Router lets a game mount handlers under two roles. The runtime decides
// what auth/wrapping to apply for each role; the game just declares its
// handler's role.
type Router interface {
	// Admin handlers require operator privilege: room CRUD, kick, list.
	Admin(pattern string, h http.Handler)
	// Public handlers serve authenticated players: WebSocket, read-only player API.
	Public(pattern string, h http.Handler)
}
