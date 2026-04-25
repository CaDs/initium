package handler

import "time"

// User is the wire-shape for the authenticated user — replaces the
// previously-generated api.User. Field tags drive Huma's spec output.
type User struct {
	ID        string    `json:"id" format:"uuid" doc:"User UUID"`
	Email     string    `json:"email" format:"email" doc:"Email address"`
	Name      string    `json:"name" doc:"Display name"`
	AvatarURL string    `json:"avatar_url" doc:"URL to avatar image (empty when not set)"`
	Role      string    `json:"role" enum:"user,admin" doc:"Authorization role"`
	CreatedAt time.Time `json:"created_at" format:"date-time" doc:"Account creation timestamp"`
}

// TokenPair is returned by every endpoint that issues a session
// (mobile/google, mobile/verify, refresh). Mirrors the old api.TokenPair.
type TokenPair struct {
	AccessToken  string `json:"access_token" doc:"Short-lived JWT (~15 min) for the Authorization: Bearer header"`
	RefreshToken string `json:"refresh_token" doc:"Long-lived token (~7 days) for renewing the access token"`
}

// MessageResponse is the generic ack wrapper for endpoints that don't
// return structured data (e.g. logout, magic-link request).
type MessageResponse struct {
	Message string `json:"message"`
}

// AdminPingResponse is the trivial "is admin role enforcement working" probe.
type AdminPingResponse struct {
	Role string `json:"role" enum:"admin"`
}
