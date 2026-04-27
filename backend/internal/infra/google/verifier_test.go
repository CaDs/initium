package google

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/domain"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestOAuthVerifier_FetchProfile_NonOK_ReturnsInvalidOAuthToken(t *testing.T) {
	t.Parallel()

	verifier := NewOAuthVerifier("client-id", "secret", "http://localhost/callback")
	verifier.profileURL = "https://example.test/userinfo"
	verifier.httpClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusUnauthorized,
			Body:       io.NopCloser(strings.NewReader(`{}`)),
			Header:     make(http.Header),
		}, nil
	})}

	_, err := verifier.fetchProfile(context.Background(), nil)

	require.ErrorIs(t, err, domain.ErrInvalidOAuthToken)
}
