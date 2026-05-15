package remote

import (
	"log"
	"net/http"
	"strings"
)

func requireAdminToken(adminToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if presented(r) != adminToken {
				log.Printf("admin: rejected %s %s (remote=%s)", r.Method, r.URL.Path, r.RemoteAddr)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func presented(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		if strings.HasPrefix(h, "Bearer ") {
			return strings.TrimPrefix(h, "Bearer ")
		}
	}
	return r.URL.Query().Get("admin")
}

func requireMobileSession(store *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := mobileToken(r)
			if token == "" || !store.Validate(token) {
				log.Printf("session: rejected %s %s (remote=%s)", r.Method, r.URL.Path, r.RemoteAddr)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func mobileToken(r *http.Request) string {
	if t := r.URL.Query().Get("t"); t != "" {
		return t
	}
	if c, err := r.Cookie("mobile_token"); err == nil {
		return c.Value
	}
	return r.Header.Get("X-Mobile-Token")
}
