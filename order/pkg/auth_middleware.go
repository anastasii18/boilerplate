package pkg

import (
	"context"
	"net/http"
)

type contextKey string

const SessionUUIDKey contextKey = "session_uuid"

func AuthMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionUuid := r.Header.Get("X-Session-Uuid")

		if sessionUuid == "" {
			http.Error(w, "session-uuid is required", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), SessionUUIDKey, sessionUuid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
