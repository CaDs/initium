package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"

	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/domain"
)

// HumaAuthMiddleware returns a Huma middleware that validates JWT access
// tokens. Mirrors middleware.Auth's chi-native behavior so handler code
// reads userID via middleware.GetUserID(ctx.Context()) unchanged.
//
// On success: stores UserIDKey + EmailKey in the request context, calls next.
// On failure: writes a 401 ErrorResponse via huma.WriteErr and stops the chain.
//
// devBypass=true short-circuits to a stub user, matching the chi version
// for parity with `DEV_BYPASS_AUTH=true` flows.
func HumaAuthMiddleware(api huma.API, tokens domain.TokenGenerator, devBypass bool) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		var userID, email string
		if devBypass {
			userID = "00000000-0000-0000-0000-000000000001"
			email = "dev@initium.local"
		} else {
			token := extractBearer(ctx)
			if token == "" {
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized,
					"missing access token", domain.ErrInvalidCredentials)
				return
			}
			var err error
			userID, email, err = tokens.ValidateAccessToken(token)
			if err != nil {
				_ = huma.WriteErr(api, ctx, http.StatusUnauthorized,
					"invalid access token", err)
				return
			}
		}

		c := context.WithValue(ctx.Context(), middleware.UserIDKey, userID)
		c = context.WithValue(c, middleware.EmailKey, email)
		next(huma.WithContext(ctx, c))
	}
}

// HumaRequireRole returns a Huma middleware that 403s if the authenticated
// user's role does not match the required role. Must run AFTER
// HumaAuthMiddleware in the operation's Middlewares slice.
func HumaRequireRole(api huma.API, role string, lookup middleware.RoleLookupFn) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		userID := middleware.GetUserID(ctx.Context())
		if userID == "" {
			_ = huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"missing user identity", domain.ErrInvalidCredentials)
			return
		}
		got, err := lookup(ctx.Context(), userID)
		if err != nil {
			_ = huma.WriteErr(api, ctx, http.StatusInternalServerError,
				"role lookup failed", err)
			return
		}
		if got != role {
			_ = huma.WriteErr(api, ctx, http.StatusForbidden,
				"insufficient role", errors.New("forbidden"))
			return
		}
		next(ctx)
	}
}

// extractBearer reads the access token from `Authorization: Bearer ...` or
// the `access_token` cookie. Mirrors the chi middleware's extractToken().
func extractBearer(ctx huma.Context) string {
	if auth := ctx.Header("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	// Parse cookies manually — huma.Context has no Cookie() method.
	if cookieHdr := ctx.Header("Cookie"); cookieHdr != "" {
		header := http.Header{}
		header.Set("Cookie", cookieHdr)
		req := http.Request{Header: header}
		if c, err := req.Cookie("access_token"); err == nil {
			return c.Value
		}
	}
	return ""
}
