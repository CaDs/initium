package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/eridia/initium/backend/internal/domain"
)

// --- Mocks ---

type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) FindByID(_ context.Context, id string) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	user.ID = "generated-id"
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	m.users[user.ID] = user
	return nil
}

type mockSessionRepo struct {
	sessions   map[string]*domain.Session
	magicLinks map[string]*domain.MagicLinkToken
	revoked    []string // tracks revoked session IDs
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions:   make(map[string]*domain.Session),
		magicLinks: make(map[string]*domain.MagicLinkToken),
	}
}

func (m *mockSessionRepo) CreateSession(_ context.Context, s *domain.Session) error {
	s.ID = fmt.Sprintf("session-%d", len(m.sessions)+1)
	s.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	s.CreatedAt = time.Now()
	m.sessions[s.RefreshTokenHash] = s
	return nil
}

func (m *mockSessionRepo) FindSessionByRefreshTokenHash(_ context.Context, hash string) (*domain.Session, error) {
	s, ok := m.sessions[hash]
	if !ok {
		return nil, domain.ErrSessionNotFound
	}
	return s, nil
}

func (m *mockSessionRepo) RevokeSession(_ context.Context, id string) error {
	m.revoked = append(m.revoked, id)
	for _, s := range m.sessions {
		if s.ID == id {
			now := time.Now()
			s.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockSessionRepo) RevokeAllUserSessions(_ context.Context, userID string) error {
	now := time.Now()
	for _, s := range m.sessions {
		if s.UserID == userID {
			s.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockSessionRepo) CreateMagicLinkToken(_ context.Context, t *domain.MagicLinkToken) error {
	t.ID = "mlt-1"
	t.ExpiresAt = time.Now().Add(15 * time.Minute)
	t.CreatedAt = time.Now()
	m.magicLinks[t.TokenHash] = t
	return nil
}

func (m *mockSessionRepo) FindMagicLinkTokenByHash(_ context.Context, hash string) (*domain.MagicLinkToken, error) {
	t, ok := m.magicLinks[hash]
	if !ok {
		return nil, domain.ErrTokenInvalid
	}
	return t, nil
}

func (m *mockSessionRepo) MarkMagicLinkTokenUsed(_ context.Context, id string) error {
	for _, t := range m.magicLinks {
		if t.ID == id {
			if t.UsedAt != nil {
				return domain.ErrTokenUsed
			}
			now := time.Now()
			t.UsedAt = &now
			return nil
		}
	}
	return nil
}

type mockOAuthVerifier struct{}

func (m *mockOAuthVerifier) ExchangeCode(_ context.Context, code string) (*domain.OAuthProfile, error) {
	if code == "valid-code" {
		return &domain.OAuthProfile{Email: "google@test.com", Name: "Google User"}, nil
	}
	return nil, domain.ErrInvalidOAuthToken
}

func (m *mockOAuthVerifier) VerifyIDToken(_ context.Context, idToken string) (*domain.OAuthProfile, error) {
	if idToken == "valid-id-token" {
		return &domain.OAuthProfile{Email: "mobile@test.com", Name: "Mobile User"}, nil
	}
	return nil, domain.ErrInvalidOAuthToken
}

type mockEmailSender struct {
	sent []string
}

func (m *mockEmailSender) SendMagicLink(_ context.Context, to, token string) error {
	m.sent = append(m.sent, to)
	return nil
}

type mockTokenGen struct{}

func (m *mockTokenGen) GenerateAccessToken(userID, email string) (string, error) {
	return "access-" + userID, nil
}

func (m *mockTokenGen) GenerateRefreshToken() (string, error) {
	return "refresh-token", nil
}

func (m *mockTokenGen) ValidateAccessToken(token string) (string, string, error) {
	return "", "", nil
}

func (m *mockTokenGen) HashToken(token string) string {
	return "hash-" + token
}

func newTestAuthService() (*AuthService, *mockUserRepo, *mockSessionRepo, *mockEmailSender) {
	users := newMockUserRepo()
	sessions := newMockSessionRepo()
	email := &mockEmailSender{}
	svc := NewAuthService(users, sessions, &mockOAuthVerifier{}, email, &mockTokenGen{})
	return svc, users, sessions, email
}

// --- Tests ---

func TestAuthService_LoginWithGoogle_CreatesNewUser(t *testing.T) {
	t.Parallel()
	svc, users, _, _ := newTestAuthService()

	user, pair, err := svc.LoginWithGoogle(context.Background(), "valid-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Email != "google@test.com" {
		t.Errorf("expected google@test.com, got %q", user.Email)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Error("expected non-empty token pair")
	}
	if len(users.users) != 1 {
		t.Errorf("expected 1 user created, got %d", len(users.users))
	}
}

func TestAuthService_LoginWithGoogle_ReturnsExistingUser(t *testing.T) {
	t.Parallel()
	svc, users, _, _ := newTestAuthService()

	users.users["existing"] = &domain.User{ID: "existing", Email: "google@test.com", Name: "Existing"}

	user, _, err := svc.LoginWithGoogle(context.Background(), "valid-code")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != "existing" {
		t.Errorf("expected existing user ID, got %q", user.ID)
	}
	if len(users.users) != 1 {
		t.Errorf("expected no new user, got %d users", len(users.users))
	}
}

func TestAuthService_LoginWithGoogle_InvalidCode(t *testing.T) {
	t.Parallel()
	svc, _, _, _ := newTestAuthService()

	_, _, err := svc.LoginWithGoogle(context.Background(), "bad-code")
	if err == nil {
		t.Fatal("expected error for invalid code")
	}
}

func TestAuthService_RequestMagicLink_SendsEmail(t *testing.T) {
	t.Parallel()
	svc, _, sessions, emailSender := newTestAuthService()

	err := svc.RequestMagicLink(context.Background(), "user@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(emailSender.sent) != 1 || emailSender.sent[0] != "user@test.com" {
		t.Errorf("expected email sent to user@test.com, got %v", emailSender.sent)
	}
	if len(sessions.magicLinks) != 1 {
		t.Errorf("expected 1 magic link token, got %d", len(sessions.magicLinks))
	}
}

func TestAuthService_RequestMagicLink_EmptyEmail(t *testing.T) {
	t.Parallel()
	svc, _, _, _ := newTestAuthService()

	err := svc.RequestMagicLink(context.Background(), "")
	if !errors.Is(err, domain.ErrEmailRequired) {
		t.Errorf("expected ErrEmailRequired, got %v", err)
	}
}

func TestAuthService_VerifyMagicLink_CreatesSession(t *testing.T) {
	t.Parallel()
	svc, _, sessions, _ := newTestAuthService()

	// Create a magic link token manually
	sessions.magicLinks["hash-test-token"] = &domain.MagicLinkToken{
		ID:        "mlt-1",
		Email:     "magic@test.com",
		TokenHash: "hash-test-token",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	user, pair, err := svc.VerifyMagicLink(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Email != "magic@test.com" {
		t.Errorf("expected magic@test.com, got %q", user.Email)
	}
	if pair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestAuthService_VerifyMagicLink_ExpiredToken(t *testing.T) {
	t.Parallel()
	svc, _, sessions, _ := newTestAuthService()

	sessions.magicLinks["hash-expired"] = &domain.MagicLinkToken{
		ID:        "mlt-2",
		Email:     "expired@test.com",
		TokenHash: "hash-expired",
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}

	_, _, err := svc.VerifyMagicLink(context.Background(), "expired")
	if !errors.Is(err, domain.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestAuthService_VerifyMagicLink_RedeemTwice_ReturnsTokenUsed(t *testing.T) {
	t.Parallel()
	svc, _, sessions, _ := newTestAuthService()

	sessions.magicLinks["hash-twice"] = &domain.MagicLinkToken{
		ID:        "mlt-twice",
		Email:     "twice@test.com",
		TokenHash: "hash-twice",
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if _, _, err := svc.VerifyMagicLink(context.Background(), "twice"); err != nil {
		t.Fatalf("first redeem should succeed, got %v", err)
	}

	_, _, err := svc.VerifyMagicLink(context.Background(), "twice")
	if !errors.Is(err, domain.ErrTokenUsed) {
		t.Errorf("second redeem should return ErrTokenUsed, got %v", err)
	}
}

func TestAuthService_VerifyMagicLink_UsedToken(t *testing.T) {
	t.Parallel()
	svc, _, sessions, _ := newTestAuthService()

	usedAt := time.Now()
	sessions.magicLinks["hash-used"] = &domain.MagicLinkToken{
		ID:        "mlt-3",
		Email:     "used@test.com",
		TokenHash: "hash-used",
		ExpiresAt: time.Now().Add(10 * time.Minute),
		UsedAt:    &usedAt,
	}

	_, _, err := svc.VerifyMagicLink(context.Background(), "used")
	if !errors.Is(err, domain.ErrTokenUsed) {
		t.Errorf("expected ErrTokenUsed, got %v", err)
	}
}

func TestAuthService_RefreshTokens_RotatesSession(t *testing.T) {
	t.Parallel()
	svc, users, sessions, _ := newTestAuthService()

	users.users["user-1"] = &domain.User{ID: "user-1", Email: "refresh@test.com"}
	sessions.sessions["hash-refresh-token"] = &domain.Session{
		ID:               "sess-1",
		UserID:           "user-1",
		RefreshTokenHash: "hash-refresh-token",
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
	}

	pair, err := svc.RefreshTokens(context.Background(), "refresh-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pair.AccessToken == "" || pair.RefreshToken == "" {
		t.Error("expected non-empty token pair")
	}

	// Old session should be revoked
	if len(sessions.revoked) != 1 || sessions.revoked[0] != "sess-1" {
		t.Errorf("expected session sess-1 revoked, got %v", sessions.revoked)
	}
}

func TestAuthService_RefreshTokens_ExpiredSession(t *testing.T) {
	t.Parallel()
	svc, _, sessions, _ := newTestAuthService()

	sessions.sessions["hash-expired-refresh"] = &domain.Session{
		ID:               "sess-2",
		UserID:           "user-1",
		RefreshTokenHash: "hash-expired-refresh",
		ExpiresAt:        time.Now().Add(-1 * time.Hour),
	}

	_, err := svc.RefreshTokens(context.Background(), "expired-refresh")
	if !errors.Is(err, domain.ErrSessionExpired) {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}

func TestAuthService_Logout_RevokesSession(t *testing.T) {
	t.Parallel()
	svc, _, sessions, _ := newTestAuthService()

	sessions.sessions["hash-logout-token"] = &domain.Session{
		ID:               "sess-3",
		UserID:           "user-1",
		RefreshTokenHash: "hash-logout-token",
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
	}

	err := svc.Logout(context.Background(), "logout-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	session := sessions.sessions["hash-logout-token"]
	if session.RevokedAt == nil {
		t.Error("expected session to be revoked")
	}
}

func TestAuthService_LogoutAll_RevokesAllSessions(t *testing.T) {
	t.Parallel()
	svc, _, sessions, _ := newTestAuthService()

	sessions.sessions["hash-1"] = &domain.Session{ID: "s1", UserID: "user-1", RefreshTokenHash: "hash-1", ExpiresAt: time.Now().Add(time.Hour)}
	sessions.sessions["hash-2"] = &domain.Session{ID: "s2", UserID: "user-1", RefreshTokenHash: "hash-2", ExpiresAt: time.Now().Add(time.Hour)}
	sessions.sessions["hash-3"] = &domain.Session{ID: "s3", UserID: "user-2", RefreshTokenHash: "hash-3", ExpiresAt: time.Now().Add(time.Hour)}

	err := svc.LogoutAll(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sessions.sessions["hash-1"].RevokedAt == nil {
		t.Error("expected session 1 to be revoked")
	}
	if sessions.sessions["hash-2"].RevokedAt == nil {
		t.Error("expected session 2 to be revoked")
	}
	if sessions.sessions["hash-3"].RevokedAt != nil {
		t.Error("expected session 3 (different user) to NOT be revoked")
	}
}

func TestAuthService_VerifyGoogleIDToken_CreatesUser(t *testing.T) {
	t.Parallel()
	svc, _, _, _ := newTestAuthService()

	user, pair, err := svc.VerifyGoogleIDToken(context.Background(), "valid-id-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Email != "mobile@test.com" {
		t.Errorf("expected mobile@test.com, got %q", user.Email)
	}
	if pair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestAuthService_FindOrCreateUser_DistinguishesDBErrors(t *testing.T) {
	t.Parallel()

	// Create a user repo that returns a non-ErrUserNotFound error
	badRepo := &errorUserRepo{err: errors.New("connection refused")}
	sessions := newMockSessionRepo()
	svc := NewAuthService(badRepo, sessions, &mockOAuthVerifier{}, &mockEmailSender{}, &mockTokenGen{})

	_, _, err := svc.LoginWithGoogle(context.Background(), "valid-code")
	if err == nil {
		t.Fatal("expected error when DB fails")
	}
	if err.Error() == "" || errors.Is(err, domain.ErrUserNotFound) {
		t.Error("error should be a DB error, not ErrUserNotFound")
	}
}

// errorUserRepo always returns the configured error from FindByEmail.
type errorUserRepo struct {
	err error
}

func (e *errorUserRepo) FindByID(_ context.Context, _ string) (*domain.User, error) {
	return nil, e.err
}
func (e *errorUserRepo) FindByEmail(_ context.Context, _ string) (*domain.User, error) {
	return nil, e.err
}
func (e *errorUserRepo) Create(_ context.Context, _ *domain.User) error { return nil }
func (e *errorUserRepo) Update(_ context.Context, _ *domain.User) error { return nil }
