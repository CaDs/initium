package token

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"
)

func setupTestGenerator(t *testing.T) *JWTGenerator {
	t.Helper()

	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generating ed25519 key: %v", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshalling private key: %v", err)
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		t.Fatalf("marshalling public key: %v", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	privFile := t.TempDir() + "/priv.pem"
	pubFile := t.TempDir() + "/pub.pem"
	os.WriteFile(privFile, privPEM, 0600)
	os.WriteFile(pubFile, pubPEM, 0644)

	gen, err := NewJWTGenerator(privFile, pubFile)
	if err != nil {
		t.Fatalf("creating JWT generator: %v", err)
	}
	return gen
}

func TestJWTGenerator_GenerateAndValidateAccessToken(t *testing.T) {
	t.Parallel()
	gen := setupTestGenerator(t)

	token, err := gen.GenerateAccessToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("generating access token: %v", err)
	}

	if token == "" {
		t.Fatal("expected non-empty token")
	}

	userID, email, err := gen.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("validating access token: %v", err)
	}

	if userID != "user-123" {
		t.Errorf("expected userID %q, got %q", "user-123", userID)
	}
	if email != "test@example.com" {
		t.Errorf("expected email %q, got %q", "test@example.com", email)
	}
}

func TestJWTGenerator_ValidateAccessToken_InvalidToken(t *testing.T) {
	t.Parallel()
	gen := setupTestGenerator(t)

	_, _, err := gen.ValidateAccessToken("not-a-jwt")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestJWTGenerator_ValidateAccessToken_WrongKey(t *testing.T) {
	t.Parallel()
	gen1 := setupTestGenerator(t)
	gen2 := setupTestGenerator(t)

	token, err := gen1.GenerateAccessToken("user-123", "test@example.com")
	if err != nil {
		t.Fatalf("generating token: %v", err)
	}

	_, _, err = gen2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error when validating with wrong key")
	}
}

func TestJWTGenerator_GenerateRefreshToken(t *testing.T) {
	t.Parallel()
	gen := setupTestGenerator(t)

	token1, err := gen.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generating refresh token: %v", err)
	}

	token2, err := gen.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generating refresh token: %v", err)
	}

	if token1 == "" || token2 == "" {
		t.Fatal("expected non-empty tokens")
	}
	if token1 == token2 {
		t.Error("expected unique refresh tokens")
	}
}

func TestJWTGenerator_HashToken(t *testing.T) {
	t.Parallel()
	gen := setupTestGenerator(t)

	hash1 := gen.HashToken("my-token")
	hash2 := gen.HashToken("my-token")
	hash3 := gen.HashToken("different-token")

	if hash1 != hash2 {
		t.Error("same input should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different inputs should produce different hashes")
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64-char hex SHA-256, got %d chars", len(hash1))
	}
}
