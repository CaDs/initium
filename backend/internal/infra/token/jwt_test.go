package token

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestGenerator(t *testing.T) *JWTGenerator {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err, "generating ed25519 key")

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	require.NoError(t, err, "marshalling private key")
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err, "marshalling public key")

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	dir := t.TempDir()
	privFile := dir + "/priv.pem"
	pubFile := dir + "/pub.pem"
	require.NoError(t, os.WriteFile(privFile, privPEM, 0600))
	require.NoError(t, os.WriteFile(pubFile, pubPEM, 0600))

	gen, err := NewJWTGenerator(privFile, pubFile)
	require.NoError(t, err, "creating JWT generator")
	return gen
}

func TestJWTGenerator_GenerateAndValidateAccessToken(t *testing.T) {
	t.Parallel()
	gen := setupTestGenerator(t)

	token, err := gen.GenerateAccessToken("user-123", "test@example.com", "admin")
	require.NoError(t, err)
	assert.NotEmpty(t, token, "token must not be empty")

	userID, email, role, err := gen.ValidateAccessToken(token)
	require.NoError(t, err)
	assert.Equal(t, "user-123", userID)
	assert.Equal(t, "test@example.com", email)
	assert.Equal(t, "admin", role)
}

func TestJWTGenerator_ValidateAccessToken_InvalidToken_ReturnsError(t *testing.T) {
	t.Parallel()
	gen := setupTestGenerator(t)

	_, _, _, err := gen.ValidateAccessToken("not-a-jwt")
	assert.Error(t, err, "expected error for invalid token")
}

func TestJWTGenerator_ValidateAccessToken_WrongKey_ReturnsError(t *testing.T) {
	t.Parallel()
	gen1 := setupTestGenerator(t)
	gen2 := setupTestGenerator(t)

	token, err := gen1.GenerateAccessToken("user-123", "test@example.com", "user")
	require.NoError(t, err)

	_, _, _, err = gen2.ValidateAccessToken(token)
	assert.Error(t, err, "validating with wrong key must fail")
}

func TestJWTGenerator_GenerateRefreshToken_ReturnsUniqueTokens(t *testing.T) {
	t.Parallel()
	gen := setupTestGenerator(t)

	token1, err := gen.GenerateRefreshToken()
	require.NoError(t, err)
	token2, err := gen.GenerateRefreshToken()
	require.NoError(t, err)

	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2, "refresh tokens must be unique")
}

func TestJWTGenerator_HashToken_DeterministicAndCollisionResistant(t *testing.T) {
	t.Parallel()
	gen := setupTestGenerator(t)

	hash1 := gen.HashToken("my-token")
	hash2 := gen.HashToken("my-token")
	hash3 := gen.HashToken("different-token")

	assert.Equal(t, hash1, hash2, "same input must produce same hash")
	assert.NotEqual(t, hash1, hash3, "different inputs must produce different hashes")
	assert.Len(t, hash1, 64, "SHA-256 hex output must be 64 characters")
}
