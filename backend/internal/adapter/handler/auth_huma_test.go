package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/testutil"
)

// HumaAuthMiddleware + HumaRequireRole + extractBearer cover every
// protected endpoint. These tests are the primary safety net for the
// auth layer — a regression here means the wrong user (or no user)
// reaches handler code.
//
// Each test registers a single tiny GET /protected operation that
// echoes back the userID middleware injected. We can then assert on
// status code (auth blocked vs allowed) AND on the userID handed
// down (auth ran correctly).

// echoHandler returns the middleware-injected userID from the request
// context so tests can verify the auth layer wired the right value.
type echoOutput struct {
	Body struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		Role   string `json:"role"`
	}
}

func registerEchoEndpoint(api huma.API, mws ...func(huma.Context, func(huma.Context))) {
	huma.Register(api, huma.Operation{
		OperationID: "echo-userid",
		Method:      http.MethodGet,
		Path:        "/protected",
		Middlewares: huma.Middlewares(mws),
	}, func(ctx context.Context, _ *struct{}) (*echoOutput, error) {
		out := &echoOutput{}
		out.Body.UserID = middleware.GetUserID(ctx)
		if email, ok := ctx.Value(middleware.EmailKey).(string); ok {
			out.Body.Email = email
		}
		out.Body.Role = middleware.GetRole(ctx)
		return out, nil
	})
}

// ----------------------------------------------------------------------
// HumaAuthMiddleware
// ----------------------------------------------------------------------

func TestHumaAuthMiddleware_DevBypass_InjectsStubUser(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	tokens := &testutil.MockTokenGenerator{}
	registerEchoEndpoint(api, handler.HumaAuthMiddleware(api, tokens, true))

	resp := api.Get("/protected")

	require.Equal(t, http.StatusOK, resp.Code)
	out := testutil.MustDecodeJSON[struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
	}](t, resp.Body)
	assert.Equal(t, "00000000-0000-0000-0000-000000000001", out.UserID)
	assert.Equal(t, "dev@initium.local", out.Email)
}

func TestHumaAuthMiddleware_ValidBearerHeader_AuthenticatesUser(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(token string) (string, string, string, error) {
			require.Equal(t, "valid-token", token)
			return "user-42", "user@example.com", "admin", nil
		},
	}
	registerEchoEndpoint(api, handler.HumaAuthMiddleware(api, tokens, false))

	resp := api.Get("/protected", "Authorization: Bearer valid-token")

	require.Equal(t, http.StatusOK, resp.Code)
	out := testutil.MustDecodeJSON[struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		Role   string `json:"role"`
	}](t, resp.Body)
	assert.Equal(t, "user-42", out.UserID)
	assert.Equal(t, "user@example.com", out.Email)
	assert.Equal(t, "admin", out.Role)
}

func TestHumaAuthMiddleware_ValidCookieToken_AuthenticatesUser(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(token string) (string, string, string, error) {
			require.Equal(t, "cookie-token", token)
			return "user-cookie", "cookie@example.com", "user", nil
		},
	}
	registerEchoEndpoint(api, handler.HumaAuthMiddleware(api, tokens, false))

	resp := api.Get("/protected", "Cookie: access_token=cookie-token")

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestHumaAuthMiddleware_BearerOverridesCookie(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(token string) (string, string, string, error) {
			// extractBearer prefers the header; cookie should be ignored.
			require.Equal(t, "header-token", token)
			return "header-user", "header@example.com", "user", nil
		},
	}
	registerEchoEndpoint(api, handler.HumaAuthMiddleware(api, tokens, false))

	resp := api.Get("/protected",
		"Authorization: Bearer header-token",
		"Cookie: access_token=cookie-token")

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestHumaAuthMiddleware_NoToken_Returns401(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	registerEchoEndpoint(api, handler.HumaAuthMiddleware(api, &testutil.MockTokenGenerator{}, false))

	resp := api.Get("/protected")

	require.Equal(t, http.StatusUnauthorized, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "INVALID_CREDENTIALS", env.Code)
}

func TestHumaAuthMiddleware_InvalidToken_Returns401(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(_ string) (string, string, string, error) {
			return "", "", "", errors.New("signature invalid")
		},
	}
	registerEchoEndpoint(api, handler.HumaAuthMiddleware(api, tokens, false))

	resp := api.Get("/protected", "Authorization: Bearer bad-token")

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

func TestHumaAuthMiddleware_NonBearerScheme_Ignored(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	registerEchoEndpoint(api, handler.HumaAuthMiddleware(api, &testutil.MockTokenGenerator{}, false))

	// "Basic" auth, not Bearer — extractBearer returns "".
	resp := api.Get("/protected", "Authorization: Basic dXNlcjpwYXNz")

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

// ----------------------------------------------------------------------
// HumaRequireRole
// ----------------------------------------------------------------------

func TestHumaRequireRole_AdminRole_Allows(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	authMW := handler.HumaAuthMiddleware(api, &testutil.MockTokenGenerator{}, true)
	roleMW := handler.HumaRequireRole(api, "admin", func(_ context.Context, userID string) (string, error) {
		t.Fatalf("lookup should not run when auth middleware already supplied role for %s", userID)
		return "", nil
	})
	registerEchoEndpoint(api, authMW, roleMW)

	resp := api.Get("/protected")

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestHumaRequireRole_WrongRole_Returns403(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(_ string) (string, string, string, error) {
			return "user-1", "u@example.com", "user", nil
		},
	}
	authMW := handler.HumaAuthMiddleware(api, tokens, false)
	roleMW := handler.HumaRequireRole(api, "admin", func(_ context.Context, _ string) (string, error) {
		t.Fatal("lookup should not run when auth middleware supplied a role")
		return "", nil
	})
	registerEchoEndpoint(api, authMW, roleMW)

	resp := api.Get("/protected", "Authorization: Bearer token")

	require.Equal(t, http.StatusForbidden, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "FORBIDDEN", env.Code)
}

func TestHumaRequireRole_LookupError_Returns500(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(_ string) (string, string, string, error) {
			return "user-1", "u@example.com", "", nil
		},
	}
	authMW := handler.HumaAuthMiddleware(api, tokens, false)
	roleMW := handler.HumaRequireRole(api, "admin", func(_ context.Context, _ string) (string, error) {
		return "", errors.New("database down")
	})
	registerEchoEndpoint(api, authMW, roleMW)

	resp := api.Get("/protected", "Authorization: Bearer token")

	require.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestHumaRequireRole_MissingUserID_Returns401(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	// Note: no auth middleware — userID never gets injected.
	roleMW := handler.HumaRequireRole(api, "admin", func(_ context.Context, _ string) (string, error) {
		t.Fatal("lookup should not be called when userID is missing")
		return "", nil
	})
	registerEchoEndpoint(api, roleMW)

	resp := api.Get("/protected")

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}

// ----------------------------------------------------------------------
// Middleware ordering: auth middleware records context state that the
// next middleware reads. Verify the chain works end-to-end.
// ----------------------------------------------------------------------

func TestHumaAuthMiddleware_WithRoleMiddleware_UsesTokenRoleWithoutLookup(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	tokens := &testutil.MockTokenGenerator{
		ValidateAccessTokenFn: func(_ string) (string, string, string, error) {
			return "user-from-token", "u@example.com", "admin", nil
		},
	}
	authMW := handler.HumaAuthMiddleware(api, tokens, false)

	roleMW := handler.HumaRequireRole(api, "admin", func(_ context.Context, userID string) (string, error) {
		t.Fatalf("lookup should not run when auth middleware supplied role for %s", userID)
		return "", nil
	})
	registerEchoEndpoint(api, authMW, roleMW)

	resp := api.Get("/protected", "Authorization: Bearer t")

	require.Equal(t, http.StatusOK, resp.Code)
}

// ----------------------------------------------------------------------
// Compile-time assertion: domain.ErrInvalidCredentials remains usable
// as the failure-case error so MapDomainErr emits the right code.
// ----------------------------------------------------------------------

var _ = domain.ErrInvalidCredentials
