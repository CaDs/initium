package testutil

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/require"
)

// MustDecodeJSON decodes an io.Reader (typically a humatest response
// body) as JSON into T, failing the test on parse error. Returns the
// decoded value so the call site can assert on fields directly.
//
//	user := testutil.MustDecodeJSON[handler.User](t, resp.Body)
//	assert.Equal(t, "admin@initium.local", user.Email)
func MustDecodeJSON[T any](t *testing.T, body io.Reader) T {
	t.Helper()
	var v T
	require.NoError(t, json.NewDecoder(body).Decode(&v),
		"decoding response body as %T", v)
	return v
}

// NoopMiddleware is a Huma middleware that just calls next. Use it
// as a stand-in for the rate limiter or any chi-bridged middleware
// in tests that don't exercise that layer.
func NoopMiddleware(ctx huma.Context, next func(huma.Context)) { next(ctx) }
