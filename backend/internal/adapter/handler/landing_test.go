package handler_test

import (
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/adapter/handler"
	"github.com/eridia/initium/backend/internal/testutil"
)

// landing_test.go is a JSON-shape smoke test. The handler returns a
// constant struct, so there's nothing to mock — the value of the test
// is asserting the wire shape stays stable. If anyone accidentally
// renames a field or drops one, web Zod schemas / mobile DTOs would
// break and this test fails first.

func TestRegisterLanding_GET_Returns200_AndExpectedShape(t *testing.T) {
	t.Parallel()
	handler.InstallErrorEnvelope()

	_, api := humatest.New(t)
	handler.RegisterLanding(api)

	resp := api.Get("/api/landing")

	require.Equal(t, http.StatusOK, resp.Code)
	info := testutil.MustDecodeJSON[handler.LandingInfo](t, resp.Body)
	assert.Equal(t, "Initium", info.Name)
	assert.NotEmpty(t, info.Description, "description must not be empty")
	assert.NotEmpty(t, info.Version, "version must not be empty")
}
