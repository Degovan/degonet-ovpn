package api

import (
	"io/fs"
	"net/http"
	"strings"
)

// StaticHandler menyajikan hasil build web UI (auth-src/web/dist) yang di-embed ke binary.
func StaticHandler(web fs.FS) http.HandlerFunc {
	files := http.FileServerFS(web)
	_, err := fs.Stat(web, "index.html")
	built := err == nil

	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeJSON(w, http.StatusNotFound, map[string]any{
				"success": false,
				"error":   "not found",
			})
			return
		}
		if !built {
			http.Error(w, "Web UI belum di-build. Jalankan `npm run build` di auth-src/web lalu build ulang binary.", http.StatusServiceUnavailable)
			return
		}

		// Nama file di /assets/ mengandung hash dari Vite, jadi aman di-cache selamanya.
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		files.ServeHTTP(w, r)
	}
}
