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
	RoleKey   contextKey = "role"
)

// Auth returns middleware that validates JWT access tokens.
// If devBypass is true, injects a hardcoded test user instead.
func Auth(tokens domain.TokenGenerator, devBypass bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if devBypass {
				ctx := context.WithValue(r.Context(), UserIDKey, "00000000-0000-0000-0000-000000000001")
				ctx = context.WithValue(ctx, EmailKey, "dev@initium.local")
				ctx = context.WithValue(ctx, RoleKey, "admin")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			tokenStr := extractToken(r)
			if tokenStr == "" {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"missing access token"}`, http.StatusUnauthorized)
				return
			}

			userID, email, role, err := tokens.ValidateAccessToken(tokenStr)
			if err != nil {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"invalid access token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, EmailKey, email)
			ctx = context.WithValue(ctx, RoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the authenticated user ID from the request context.
func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(UserIDKey).(string)
	return id
}

// GetRole extracts the authenticated user's role from the request context.
func GetRole(ctx context.Context) string {
	role, _ := ctx.Value(RoleKey).(string)
	return role
}

// RoleLookupFn is a callback that resolves the role for a given user ID.
// Using a callback avoids a direct dependency on the UserRepository interface
// (which would create a tighter coupling between middleware and persistence).
type RoleLookupFn func(ctx context.Context, userID string) (role string, err error)

// RequireRole returns middleware that 403s if the authenticated user's role
// does not match the required role. Must be placed after the Auth middleware.
func RequireRole(role string, lookup RoleLookupFn) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, `{"code":"UNAUTHORIZED","message":"missing user identity"}`, http.StatusUnauthorized)
				return
			}

			userRole := GetRole(r.Context())
			if userRole == "" && lookup != nil {
				var err error
				userRole, err = lookup(r.Context(), userID)
				if err != nil {
					http.Error(w, `{"code":"INTERNAL_ERROR","message":"internal error"}`, http.StatusInternalServerError)
					return
				}
			}

			if userRole != role {
				http.Error(w, `{"code":"FORBIDDEN","message":"insufficient role"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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
