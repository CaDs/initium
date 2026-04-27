package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	googleOAuth "golang.org/x/oauth2/google"

	"github.com/eridia/initium/backend/internal/domain"
)

// OAuthVerifier implements domain.OAuthVerifier for Google.
type OAuthVerifier struct {
	config       *oauth2.Config
	httpClient   *http.Client
	tokenInfoURL string
	profileURL   string
}

// NewOAuthVerifier creates a Google OAuth verifier.
func NewOAuthVerifier(clientID, clientSecret, redirectURL string) *OAuthVerifier {
	const (
		//nolint:gosec // Public Google OAuth metadata endpoint, not a credential.
		googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"
		googleProfileURL   = "https://www.googleapis.com/oauth2/v2/userinfo"
	)

	return &OAuthVerifier{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     googleOAuth.Endpoint,
		},
		httpClient:   http.DefaultClient,
		tokenInfoURL: googleTokenInfoURL,
		profileURL:   googleProfileURL,
	}
}

// AuthCodeURL returns the Google consent screen URL with the given state.
func (v *OAuthVerifier) AuthCodeURL(state string) string {
	return v.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for a user profile.
func (v *OAuthVerifier) ExchangeCode(ctx context.Context, code string) (*domain.OAuthProfile, error) {
	token, err := v.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	return v.fetchProfile(ctx, token)
}

// VerifyIDToken verifies a Google ID token (from mobile) and returns the profile.
func (v *OAuthVerifier) VerifyIDToken(ctx context.Context, idToken string) (*domain.OAuthProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		v.tokenInfoURL+"?id_token="+url.QueryEscape(idToken), nil)
	if err != nil {
		return nil, fmt.Errorf("creating token verification request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verifying id token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.ErrInvalidOAuthToken
	}

	var info struct {
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Aud           string `json:"aud"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding token info: %w", err)
	}

	if info.Aud != v.config.ClientID {
		return nil, domain.ErrInvalidOAuthToken
	}

	if info.EmailVerified != "true" {
		return nil, domain.ErrInvalidOAuthToken
	}

	return &domain.OAuthProfile{
		Email:     info.Email,
		Name:      info.Name,
		AvatarURL: info.Picture,
	}, nil
}

func (v *OAuthVerifier) fetchProfile(ctx context.Context, token *oauth2.Token) (*domain.OAuthProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.profileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating user info request: %w", err)
	}
	if token != nil && token.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	}

	client := v.httpClient
	if token != nil {
		client = v.config.Client(ctx, token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, domain.ErrInvalidOAuthToken
	}

	var info struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decoding user info: %w", err)
	}

	return &domain.OAuthProfile{
		Email:     info.Email,
		Name:      info.Name,
		AvatarURL: info.Picture,
	}, nil
}
