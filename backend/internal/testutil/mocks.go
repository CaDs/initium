// Package testutil provides shared test infrastructure for the backend:
// canonical hand-rolled mocks, JSON decode helpers, and stable fixture
// data. Lives at internal/testutil so any package's *_test.go files can
// import it without a cycle (testutil only depends on domain).
//
// **Mock convention: function-injection.** Each mock holds one nullable
// `Fn` field per interface method. Tests set only the methods they
// exercise; unset methods return zero values. This keeps test setup
// terse — table-driven tests can override a single field per case
// without redefining the whole mock.
//
// **Single source of truth.** All mocks for a given interface live HERE,
// not duplicated across test files. Adding a method to a domain
// interface fails compilation in this file first; updating the mock
// here propagates the new signature to every test that uses it.
package testutil

import (
	"context"

	"github.com/eridia/initium/backend/internal/domain"
)

// ----------------------------------------------------------------------
// Service mocks
// ----------------------------------------------------------------------

// MockAuthService is a hand-rolled mock for domain.AuthService.
//
// Compile-time assertion that this type satisfies the interface lives at
// the bottom of the file — adding a method to AuthService fails the
// build until the mock catches up.
type MockAuthService struct {
	LoginWithGoogleFn      func(ctx context.Context, code string) (*domain.User, *domain.TokenPair, error)
	VerifyGoogleIDTokenFn  func(ctx context.Context, idToken string) (*domain.User, *domain.TokenPair, error)
	RequestMagicLinkFn     func(ctx context.Context, email string) error
	VerifyMagicLinkFn      func(ctx context.Context, token string) (*domain.User, *domain.TokenPair, error)
	RefreshTokensFn        func(ctx context.Context, refreshToken string) (*domain.TokenPair, error)
	LogoutFn               func(ctx context.Context, refreshToken string) error
	LogoutAllFn            func(ctx context.Context, userID string) error
}

func (m *MockAuthService) LoginWithGoogle(ctx context.Context, code string) (*domain.User, *domain.TokenPair, error) {
	if m.LoginWithGoogleFn != nil {
		return m.LoginWithGoogleFn(ctx, code)
	}
	return nil, nil, nil
}

func (m *MockAuthService) VerifyGoogleIDToken(ctx context.Context, idToken string) (*domain.User, *domain.TokenPair, error) {
	if m.VerifyGoogleIDTokenFn != nil {
		return m.VerifyGoogleIDTokenFn(ctx, idToken)
	}
	return nil, nil, nil
}

func (m *MockAuthService) RequestMagicLink(ctx context.Context, email string) error {
	if m.RequestMagicLinkFn != nil {
		return m.RequestMagicLinkFn(ctx, email)
	}
	return nil
}

func (m *MockAuthService) VerifyMagicLink(ctx context.Context, token string) (*domain.User, *domain.TokenPair, error) {
	if m.VerifyMagicLinkFn != nil {
		return m.VerifyMagicLinkFn(ctx, token)
	}
	return nil, nil, nil
}

func (m *MockAuthService) RefreshTokens(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	if m.RefreshTokensFn != nil {
		return m.RefreshTokensFn(ctx, refreshToken)
	}
	return nil, nil
}

func (m *MockAuthService) Logout(ctx context.Context, refreshToken string) error {
	if m.LogoutFn != nil {
		return m.LogoutFn(ctx, refreshToken)
	}
	return nil
}

func (m *MockAuthService) LogoutAll(ctx context.Context, userID string) error {
	if m.LogoutAllFn != nil {
		return m.LogoutAllFn(ctx, userID)
	}
	return nil
}

// MockUserService is a hand-rolled mock for domain.UserService.
type MockUserService struct {
	GetProfileFn    func(ctx context.Context, userID string) (*domain.User, error)
	UpdateProfileFn func(ctx context.Context, userID string, name string) (*domain.User, error)
}

func (m *MockUserService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	if m.GetProfileFn != nil {
		return m.GetProfileFn(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID string, name string) (*domain.User, error) {
	if m.UpdateProfileFn != nil {
		return m.UpdateProfileFn(ctx, userID, name)
	}
	return nil, nil
}

// ----------------------------------------------------------------------
// Repository mocks
// ----------------------------------------------------------------------

// MockUserRepository is a hand-rolled mock for domain.UserRepository.
type MockUserRepository struct {
	FindByIDFn    func(ctx context.Context, id string) (*domain.User, error)
	FindByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	CreateFn      func(ctx context.Context, user *domain.User) error
	UpdateFn      func(ctx context.Context, user *domain.User) error
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.FindByEmailFn != nil {
		return m.FindByEmailFn(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, user)
	}
	return nil
}

// MockSessionRepository is a hand-rolled mock for domain.SessionRepository.
type MockSessionRepository struct {
	CreateSessionFn                 func(ctx context.Context, session *domain.Session) error
	FindSessionByRefreshTokenHashFn func(ctx context.Context, hash string) (*domain.Session, error)
	RevokeSessionFn                 func(ctx context.Context, sessionID string) error
	RevokeAllUserSessionsFn         func(ctx context.Context, userID string) error
	CreateMagicLinkTokenFn          func(ctx context.Context, token *domain.MagicLinkToken) error
	FindMagicLinkTokenByHashFn      func(ctx context.Context, hash string) (*domain.MagicLinkToken, error)
	MarkMagicLinkTokenUsedFn        func(ctx context.Context, tokenID string) error
	DeleteExpiredMagicLinksFn       func(ctx context.Context) (int, error)
}

func (m *MockSessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	if m.CreateSessionFn != nil {
		return m.CreateSessionFn(ctx, session)
	}
	return nil
}

func (m *MockSessionRepository) FindSessionByRefreshTokenHash(ctx context.Context, hash string) (*domain.Session, error) {
	if m.FindSessionByRefreshTokenHashFn != nil {
		return m.FindSessionByRefreshTokenHashFn(ctx, hash)
	}
	return nil, nil
}

func (m *MockSessionRepository) RevokeSession(ctx context.Context, sessionID string) error {
	if m.RevokeSessionFn != nil {
		return m.RevokeSessionFn(ctx, sessionID)
	}
	return nil
}

func (m *MockSessionRepository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	if m.RevokeAllUserSessionsFn != nil {
		return m.RevokeAllUserSessionsFn(ctx, userID)
	}
	return nil
}

func (m *MockSessionRepository) CreateMagicLinkToken(ctx context.Context, token *domain.MagicLinkToken) error {
	if m.CreateMagicLinkTokenFn != nil {
		return m.CreateMagicLinkTokenFn(ctx, token)
	}
	return nil
}

func (m *MockSessionRepository) FindMagicLinkTokenByHash(ctx context.Context, hash string) (*domain.MagicLinkToken, error) {
	if m.FindMagicLinkTokenByHashFn != nil {
		return m.FindMagicLinkTokenByHashFn(ctx, hash)
	}
	return nil, nil
}

func (m *MockSessionRepository) MarkMagicLinkTokenUsed(ctx context.Context, tokenID string) error {
	if m.MarkMagicLinkTokenUsedFn != nil {
		return m.MarkMagicLinkTokenUsedFn(ctx, tokenID)
	}
	return nil
}

func (m *MockSessionRepository) DeleteExpiredMagicLinks(ctx context.Context) (int, error) {
	if m.DeleteExpiredMagicLinksFn != nil {
		return m.DeleteExpiredMagicLinksFn(ctx)
	}
	return 0, nil
}

// ----------------------------------------------------------------------
// Infrastructure port mocks
// ----------------------------------------------------------------------

// MockTokenGenerator is a hand-rolled mock for domain.TokenGenerator.
//
// Default behavior returns predictable strings ("access-{userID}",
// "refresh-{n}") so assertions don't need to predict random output.
// Override the Fn fields when a test needs specific values or errors.
type MockTokenGenerator struct {
	GenerateAccessTokenFn  func(userID string, email string) (string, error)
	GenerateRefreshTokenFn func() (string, error)
	ValidateAccessTokenFn  func(token string) (userID string, email string, err error)
	HashTokenFn            func(token string) string
}

func (m *MockTokenGenerator) GenerateAccessToken(userID string, email string) (string, error) {
	if m.GenerateAccessTokenFn != nil {
		return m.GenerateAccessTokenFn(userID, email)
	}
	return "access-" + userID, nil
}

func (m *MockTokenGenerator) GenerateRefreshToken() (string, error) {
	if m.GenerateRefreshTokenFn != nil {
		return m.GenerateRefreshTokenFn()
	}
	return "refresh-token", nil
}

func (m *MockTokenGenerator) ValidateAccessToken(token string) (string, string, error) {
	if m.ValidateAccessTokenFn != nil {
		return m.ValidateAccessTokenFn(token)
	}
	return "", "", nil
}

func (m *MockTokenGenerator) HashToken(token string) string {
	if m.HashTokenFn != nil {
		return m.HashTokenFn(token)
	}
	return "hash-" + token
}

// MockOAuthVerifier is a hand-rolled mock for domain.OAuthVerifier.
type MockOAuthVerifier struct {
	ExchangeCodeFn   func(ctx context.Context, code string) (*domain.OAuthProfile, error)
	VerifyIDTokenFn  func(ctx context.Context, idToken string) (*domain.OAuthProfile, error)
}

func (m *MockOAuthVerifier) ExchangeCode(ctx context.Context, code string) (*domain.OAuthProfile, error) {
	if m.ExchangeCodeFn != nil {
		return m.ExchangeCodeFn(ctx, code)
	}
	return nil, nil
}

func (m *MockOAuthVerifier) VerifyIDToken(ctx context.Context, idToken string) (*domain.OAuthProfile, error) {
	if m.VerifyIDTokenFn != nil {
		return m.VerifyIDTokenFn(ctx, idToken)
	}
	return nil, nil
}

// MockEmailSender is a hand-rolled mock for domain.EmailSender.
type MockEmailSender struct {
	SendMagicLinkFn func(ctx context.Context, to string, token string) error
}

func (m *MockEmailSender) SendMagicLink(ctx context.Context, to string, token string) error {
	if m.SendMagicLinkFn != nil {
		return m.SendMagicLinkFn(ctx, to, token)
	}
	return nil
}

// ----------------------------------------------------------------------
// Compile-time interface conformance assertions.
// Adding a method to any domain interface above breaks the build here
// first — fix the mock, then everywhere it's used.
// ----------------------------------------------------------------------

var (
	_ domain.AuthService       = (*MockAuthService)(nil)
	_ domain.UserService       = (*MockUserService)(nil)
	_ domain.UserRepository    = (*MockUserRepository)(nil)
	_ domain.SessionRepository = (*MockSessionRepository)(nil)
	_ domain.TokenGenerator    = (*MockTokenGenerator)(nil)
	_ domain.OAuthVerifier     = (*MockOAuthVerifier)(nil)
	_ domain.EmailSender       = (*MockEmailSender)(nil)
)
