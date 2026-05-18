package utils

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// GetIP extracts the real client IP from the request, respecting common
// reverse-proxy headers. Precedence: X-Forwarded-For → X-Real-IP → RemoteAddr.
func GetIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For may be a comma-separated list; the first entry is the client.
		return strings.TrimSpace(strings.Split(ip, ",")[0])
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	// Strip the port from RemoteAddr ("1.2.3.4:5678" → "1.2.3.4").
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// GetParam extracts a named URL path variable from a gorilla/mux request.
// Returns an empty string if the variable is not present.
func GetParam(r *http.Request, key string) string {
	return mux.Vars(r)[key]
}
