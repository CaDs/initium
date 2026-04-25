package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/testutil"
)

// UserHandler endpoints (`getMe`, `updateMe`) are tested with humatest.
// The auth middleware is replaced with NoopMiddleware here — these tests
// exercise the handler logic, not the auth layer (covered in
// auth_huma_test.go). userID propagation through the auth-attached
// context is verified end-to-end in TestHumaAuthMiddleware_With...

// newUserAPI builds a humatest API with the user endpoints registered
// behind a no-op auth middleware. Returns the test API + the underlying
// MockUserService for per-test wiring.
func newUserAPI(t *testing.T) (humatest.TestAPI, *testutil.MockUserService) {
	t.Helper()
	handler.InstallErrorEnvelope()
	svc := &testutil.MockUserService{}
	_, api := humatest.New(t)
	handler.NewUserHandler(svc).RegisterUser(api, testutil.NoopMiddleware)
	return api, svc
}

// ----------------------------------------------------------------------
// getMe
// ----------------------------------------------------------------------

func TestUserHandler_GetMe_Success_ReturnsUser(t *testing.T) {
	t.Parallel()

	api, svc := newUserAPI(t)
	svc.GetProfileFn = func(_ context.Context, _ string) (*domain.User, error) {
		return &testutil.RegularUser, nil
	}

	resp := api.Get("/api/me")

	require.Equal(t, http.StatusOK, resp.Code)
	user := testutil.MustDecodeJSON[handler.User](t, resp.Body)
	assert.Equal(t, testutil.RegularUser.ID, user.ID)
	assert.Equal(t, testutil.RegularUser.Email, user.Email)
	assert.Equal(t, testutil.RegularUser.Role, user.Role)
}

func TestUserHandler_GetMe_UserNotFound_404(t *testing.T) {
	t.Parallel()

	api, svc := newUserAPI(t)
	svc.GetProfileFn = func(_ context.Context, _ string) (*domain.User, error) {
		return nil, domain.ErrUserNotFound
	}

	resp := api.Get("/api/me")

	require.Equal(t, http.StatusNotFound, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "USER_NOT_FOUND", env.Code)
}

// ----------------------------------------------------------------------
// updateMe
// ----------------------------------------------------------------------

func TestUserHandler_UpdateMe_Success(t *testing.T) {
	t.Parallel()

	api, svc := newUserAPI(t)
	var capturedName string
	svc.UpdateProfileFn = func(_ context.Context, _ string, name string) (*domain.User, error) {
		capturedName = name
		updated := testutil.RegularUser
		updated.Name = name
		return &updated, nil
	}

	resp := api.Patch("/api/me", map[string]string{"name": "New Name"})

	require.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "New Name", capturedName)
	user := testutil.MustDecodeJSON[handler.User](t, resp.Body)
	assert.Equal(t, "New Name", user.Name)
}

func TestUserHandler_UpdateMe_EmptyName_422(t *testing.T) {
	t.Parallel()

	api, _ := newUserAPI(t)

	// minLength=1 — empty string fails validation.
	resp := api.Patch("/api/me", map[string]string{"name": ""})

	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "INVALID_INPUT", env.Code)
}

func TestUserHandler_UpdateMe_OverlyLongName_422(t *testing.T) {
	t.Parallel()

	api, _ := newUserAPI(t)

	// maxLength=100 — 101 chars fails validation.
	tooLong := make([]byte, 101)
	for i := range tooLong {
		tooLong[i] = 'a'
	}
	resp := api.Patch("/api/me", map[string]string{"name": string(tooLong)})

	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
}

func TestUserHandler_UpdateMe_ServiceError_MapsDomainErr(t *testing.T) {
	t.Parallel()

	api, svc := newUserAPI(t)
	svc.UpdateProfileFn = func(_ context.Context, _ string, _ string) (*domain.User, error) {
		return nil, domain.ErrUserNotFound
	}

	resp := api.Patch("/api/me", map[string]string{"name": "valid"})

	require.Equal(t, http.StatusNotFound, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "USER_NOT_FOUND", env.Code)
}
