package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrSessionNotFound    = errors.New("session not found")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionRevoked     = errors.New("session revoked")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenUsed          = errors.New("token already used")
	ErrTokenInvalid       = errors.New("token invalid")
	ErrInvalidOAuthToken  = errors.New("invalid oauth token")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailRequired      = errors.New("email is required")
	ErrRateLimited        = errors.New("rate limited")
	ErrInvalidInput       = errors.New("invalid input")
)
