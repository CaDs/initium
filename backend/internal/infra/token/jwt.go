package token

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenExpiry  = 15 * time.Minute
	refreshTokenLength = 32
)

// JWTGenerator implements domain.TokenGenerator using Ed25519.
type JWTGenerator struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

// NewJWTGenerator loads Ed25519 keys from PEM files.
func NewJWTGenerator(privateKeyPath, publicKeyPath string) (*JWTGenerator, error) {
	privPEM, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading private key: %w", err)
	}

	privKey, err := jwt.ParseEdPrivateKeyFromPEM(privPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing private key: %w", err)
	}

	pubPEM, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading public key: %w", err)
	}

	pubKey, err := jwt.ParseEdPublicKeyFromPEM(pubPEM)
	if err != nil {
		return nil, fmt.Errorf("parsing public key: %w", err)
	}

	return &JWTGenerator{
		privateKey: privKey.(ed25519.PrivateKey),
		publicKey:  pubKey.(ed25519.PublicKey),
	}, nil
}

// GenerateAccessToken creates a signed JWT with user claims.
func (g *JWTGenerator) GenerateAccessToken(userID string, email string, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"role":  role,
		"exp":   now.Add(accessTokenExpiry).Unix(),
		"iat":   now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	signed, err := token.SignedString(g.privateKey)
	if err != nil {
		return "", fmt.Errorf("signing access token: %w", err)
	}

	return signed, nil
}

// GenerateRefreshToken creates a cryptographically random token.
func (g *JWTGenerator) GenerateRefreshToken() (string, error) {
	b := make([]byte, refreshTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// ValidateAccessToken verifies a JWT and returns the user ID, email, and role.
func (g *JWTGenerator) ValidateAccessToken(tokenString string) (string, string, string, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return g.publicKey, nil
	})
	if err != nil {
		return "", "", "", fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", "", fmt.Errorf("invalid token claims")
	}

	userID, _ := claims.GetSubject()
	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	return userID, email, role, nil
}

// HashToken returns a SHA-256 hex digest of the given token.
func (g *JWTGenerator) HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
