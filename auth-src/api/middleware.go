package api

import (
	"net/http"
)

func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			if key == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]any{
					"success": false,
					"error":   "missing X-API-Key header",
				})
				return
			}
			if key != apiKey {
				writeJSON(w, http.StatusForbidden, map[string]any{
					"success": false,
					"error":   "invalid API key",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
