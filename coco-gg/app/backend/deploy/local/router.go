package local

import "net/http"

// router is the local-mode runtime.Router. Both Admin and Public roles
// mount handlers bare on the shared mux. The plugin binds 127.0.0.1, and
// the Wails host applies the actual auth (mobile-token, loopback-only) at
// its LAN gateway. There is no in-process auth.
type router struct{ mux *http.ServeMux }

func (r *router) Admin(pattern string, h http.Handler)  { r.mux.Handle(pattern, h) }
func (r *router) Public(pattern string, h http.Handler) { r.mux.Handle(pattern, h) }
