package remote

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func uiHandler(uiDir, urlPrefix string) http.HandlerFunc {
	fileServer := http.FileServer(http.Dir(uiDir))
	stripPrefix := strings.TrimSuffix(urlPrefix, "/")
	indexPath := filepath.Join(uiDir, "index.html")
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, stripPrefix)
		// SPA fallback for the prefix root or unknown paths: serve index.html
		// directly. We DON'T rewrite r.URL.Path to "/index.html" + FileServer,
		// because Go's FileServer auto-301s any request for a path ending in
		// "/index.html" → "./", which causes an infinite redirect loop here.
		if path == "" || path == "/" {
			http.ServeFile(w, r, indexPath)
			return
		}
		abs := filepath.Join(uiDir, path)
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			r.URL.Path = path
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, indexPath)
	}
}
