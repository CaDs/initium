package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/eridia/initium/backend/internal/domain"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	EmailKey  contextKey = "email"
)

// Auth returns middleware that validates JWT access tokens.
// If devBypass is true, injects a hardcoded test user instead.
func Auth(tokens domain.TokenGenerator, devBypass bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if devBypass {
				ctx := context.WithValue(r.Context(), UserIDKey, "00000000-0000-0000-0000-000000000001")
				ctx = context.WithValue(ctx, EmailKey, "dev@initium.local")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			tokenStr := extractToken(r)
			if tokenStr == "" {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"missing access token"}`, http.StatusUnauthorized)
				return
			}

			userID, email, err := tokens.ValidateAccessToken(tokenStr)
			if err != nil {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"invalid access token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, EmailKey, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the authenticated user ID from the request context.
func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}

func extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	if c, err := r.Cookie("access_token"); err == nil {
		return c.Value
	}
	return ""
}
