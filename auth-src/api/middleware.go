package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	sessionCookie = "degonet_session"
	sessionTTL    = 12 * time.Hour
)

// APIKeyAuth menerima header X-API-Key, atau cookie session dari login web UI.
func APIKeyAuth(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if validSession(r, apiKey) {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if key == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]any{
					"success": false,
					"error":   "missing X-API-Key header",
				})
				return
			}
			if !keyMatches(key, apiKey) {
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

func keyMatches(given, apiKey string) bool {
	return subtle.ConstantTimeCompare([]byte(given), []byte(apiKey)) == 1
}

// Nilai session: "<unix expiry>.<HMAC-SHA256(apiKey, expiry)>". Mengganti API_KEY
// otomatis membatalkan semua session.
func signSession(expiry int64, apiKey string) string {
	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write([]byte(strconv.FormatInt(expiry, 10)))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func newSessionCookie(r *http.Request, apiKey string) *http.Cookie {
	expiry := time.Now().Add(sessionTTL)
	return &http.Cookie{
		Name:     sessionCookie,
		Value:    strconv.FormatInt(expiry.Unix(), 10) + "." + signSession(expiry.Unix(), apiKey),
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteStrictMode,
	}
}

func clearSessionCookie(r *http.Request) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteStrictMode,
	}
}

func validSession(r *http.Request, apiKey string) bool {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return false
	}
	expStr, sig, ok := strings.Cut(c.Value, ".")
	if !ok {
		return false
	}
	expiry, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().Unix() > expiry {
		return false
	}
	return hmac.Equal([]byte(sig), []byte(signSession(expiry, apiKey)))
}

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}
