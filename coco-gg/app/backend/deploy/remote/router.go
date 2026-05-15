package remote

import "net/http"

type router struct {
	mux        *http.ServeMux
	wrapAdmin  func(http.Handler) http.Handler
	wrapPublic func(http.Handler) http.Handler
}

func (r *router) Admin(pattern string, h http.Handler)  { r.mux.Handle(pattern, r.wrapAdmin(h)) }
func (r *router) Public(pattern string, h http.Handler) { r.mux.Handle(pattern, r.wrapPublic(h)) }
