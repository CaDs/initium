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

// MapDomainErr is the bridge between domain sentinel errors and the
// HTTP wire envelope. Two behaviors that matter:
//   1. RequestID is propagated from the request context (set by the
//      RequestID middleware) into the response envelope.
//   2. INTERNAL_ERROR scrubs the original message — internals never
//      leak to clients.

func TestMapDomainErr_PropagatesRequestID(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "test-req-123")
	env := handler.MapDomainErr(ctx, domain.ErrEmailRequired)

	require.NotNil(t, env)
	assert.Equal(t, "EMAIL_REQUIRED", env.Code)
	assert.Equal(t, http.StatusBadRequest, env.HTTPStatus)
	assert.Equal(t, "test-req-123", env.RequestID)
	assert.NotEmpty(t, env.Message)
}

func TestMapDomainErr_NoRequestID_LeavesEmpty(t *testing.T) {
	t.Parallel()

	env := handler.MapDomainErr(context.Background(), domain.ErrEmailRequired)

	require.NotNil(t, env)
	assert.Empty(t, env.RequestID,
		"request_id should be empty when context has none")
}

func TestMapDomainErr_InternalError_ScrubsMessage(t *testing.T) {
	t.Parallel()

	internalErr := errors.New("postgres: connection refused at 10.0.0.5:5432")
	env := handler.MapDomainErr(context.Background(), internalErr)

	require.NotNil(t, env)
	assert.Equal(t, "INTERNAL_ERROR", env.Code)
	assert.Equal(t, http.StatusInternalServerError, env.HTTPStatus)
	assert.Equal(t, "internal error", env.Message,
		"the original error message must not leak to the client")
	assert.NotContains(t, env.Message, "postgres")
	assert.NotContains(t, env.Message, "10.0.0.5")
}

func TestMapDomainErr_KnownDomainError_PreservesMessage(t *testing.T) {
	t.Parallel()

	env := handler.MapDomainErr(context.Background(), domain.ErrUserNotFound)

	require.NotNil(t, env)
	assert.Equal(t, "USER_NOT_FOUND", env.Code)
	assert.Equal(t, http.StatusNotFound, env.HTTPStatus)
	// Domain error message is fine to surface — these are user-safe strings.
	assert.Equal(t, domain.ErrUserNotFound.Error(), env.Message)
}

// ----------------------------------------------------------------------
// InstallErrorEnvelope: ensure huma.NewError-synthesized errors
// (validation failures, missing body) are also formatted as our
// APIError envelope so the wire shape stays consistent across paths.
// ----------------------------------------------------------------------

func TestInstallErrorEnvelope_HumaValidationFailure_UsesAPIError(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	type body struct {
		Body struct {
			Email string `json:"email" required:"true" format:"email"`
		}
	}

	_, api := humatest.New(t)
	huma.Register(api, huma.Operation{
		OperationID: "validate-test",
		Method:      http.MethodPost,
		Path:        "/probe",
	}, func(_ context.Context, _ *body) (*struct{}, error) {
		return &struct{}{}, nil
	})

	// Empty body — validation rejects with 422.
	resp := api.Post("/probe", map[string]any{})

	require.Equal(t, http.StatusUnprocessableEntity, resp.Code)
	env := testutil.MustDecodeJSON[handler.APIError](t, resp.Body)
	assert.Equal(t, "INVALID_INPUT", env.Code,
		"422 from struct-tag validation should map to INVALID_INPUT")
	assert.NotEmpty(t, env.Message)
}

// APIError implements huma.StatusError so handlers can `return nil, err`
// directly. Test the contract.
func TestAPIError_ImplementsStatusError(t *testing.T) {
	t.Parallel()

	e := &handler.APIError{HTTPStatus: 418, Code: "TEAPOT", Message: "no coffee"}

	var statusErr huma.StatusError
	require.ErrorAs(t, e, &statusErr)
	assert.Equal(t, 418, statusErr.GetStatus())
	assert.Equal(t, "no coffee", statusErr.Error())
}
