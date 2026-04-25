package testutil

import (
	"time"

	"github.com/eridia/initium/backend/internal/domain"
)

// Stable fixture timestamps. Tests that compare CreatedAt / UpdatedAt
// against assertions don't have to predict `time.Now()`. Frozen at a
// deterministic instant so test output is reproducible.
var (
	FixtureCreatedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	FixtureUpdatedAt = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
)

// AdminUser is a canonical admin-role test user. Use when you need to
// assert role-gated paths (admin/ping). The UUID is stable across
// runs so multiple tests can co-exist.
var AdminUser = domain.User{
	ID:           "00000000-0000-0000-0000-00000000a1d1",
	Email:        "admin@initium.local",
	Name:         "Admin User",
	AvatarURL:    "",
	AuthProvider: "magic_link",
	Role:         "admin",
	CreatedAt:    FixtureCreatedAt,
	UpdatedAt:    FixtureUpdatedAt,
}

// RegularUser is a canonical non-admin test user. Use for any path
// that doesn't require admin role. The UUID matches the dev-bypass
// user the chi + Huma auth middlewares emit, so tests that hit
// `DEV_BYPASS_AUTH=true` flows can reuse this fixture.
var RegularUser = domain.User{
	ID:           "00000000-0000-0000-0000-000000000001",
	Email:        "dev@initium.local",
	Name:         "Dev User",
	AvatarURL:    "",
	AuthProvider: "magic_link",
	Role:         "user",
	CreatedAt:    FixtureCreatedAt,
	UpdatedAt:    FixtureUpdatedAt,
}

// ValidTokenPair is the canonical "successful authentication" return
// value. Use as the success path output for AuthService mock methods.
var ValidTokenPair = domain.TokenPair{
	AccessToken:  "test-access-token",
	RefreshToken: "test-refresh-token",
}
