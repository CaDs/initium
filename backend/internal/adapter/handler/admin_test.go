package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/adapter/middleware"
	"github.com/eridia/initium/backend/internal/testutil"
)

// admin_test.go exercises the admin endpoint's role enforcement.
// The endpoint itself is trivial — it returns {"role":"admin"} on
// success — so the value of these tests is verifying that the role
// middleware actually blocks non-admins (regression net for the
// "didn't enforce role" failure mode).

func TestRegisterAdmin_AdminRole_Returns200(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)

	// Real role middleware backed by a lookup that returns "admin"
	// for any userID. Combined with NoopMiddleware in the auth slot —
	// HumaRequireRole reads userID from ctx, but NoopMiddleware doesn't
	// inject one, so we need a stub that does.
	authMW := func(ctx huma.Context, next func(huma.Context)) {
		c := context.WithValue(ctx.Context(), middleware.UserIDKey, testutil.AdminUser.ID)
		next(huma.WithContext(ctx, c))
	}
	roleMW := handler.HumaRequireRole(api, "admin",
		func(_ context.Context, _ string) (string, error) {
			return "admin", nil
		})
	handler.RegisterAdmin(api, authMW, roleMW)

	resp := api.Get("/api/admin/ping")

	require.Equal(t, http.StatusOK, resp.Code)
	body := testutil.MustDecodeJSON[handler.AdminPingResponse](t, resp.Body)
	assert.Equal(t, "admin", body.Role)
}

func TestRegisterAdmin_NonAdminRole_Returns403(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)

	authMW := func(ctx huma.Context, next func(huma.Context)) {
		c := context.WithValue(ctx.Context(), middleware.UserIDKey, testutil.RegularUser.ID)
		next(huma.WithContext(ctx, c))
	}
	roleMW := handler.HumaRequireRole(api, "admin",
		func(_ context.Context, _ string) (string, error) {
			return "user", nil // wrong role
		})
	handler.RegisterAdmin(api, authMW, roleMW)

	resp := api.Get("/api/admin/ping")

	require.Equal(t, http.StatusForbidden, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "FORBIDDEN", env.Code)
}

func TestRegisterAdmin_MissingUserID_Returns401(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)

	// No auth middleware injects userID — request fails at role guard.
	roleMW := handler.HumaRequireRole(api, "admin",
		func(_ context.Context, _ string) (string, error) {
			t.Fatal("role lookup should not be called when userID is missing")
			return "", nil
		})
	handler.RegisterAdmin(api, testutil.NoopMiddleware, roleMW)

	resp := api.Get("/api/admin/ping")

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}
