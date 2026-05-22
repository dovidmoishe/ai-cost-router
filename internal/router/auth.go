package router

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

var internalServiceToken string

func loadInternalServiceToken() {
	internalServiceToken = strings.TrimSpace(os.Getenv("INTERNAL_SERVICE_TOKEN"))
}

// InternalServiceAuth requires Authorization: Bearer <INTERNAL_SERVICE_TOKEN>.
func InternalServiceAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if internalServiceToken == "" {
			writeError(w, http.StatusUnauthorized, "unauthorized", "internal service auth is not configured")
			return
		}

		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "missing or invalid Authorization header")
			return
		}

		if subtle.ConstantTimeCompare([]byte(token), []byte(internalServiceToken)) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized", "invalid authorization token")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}
