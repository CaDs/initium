package domain

import "time"

// User represents a registered user in the system.
type User struct {
	ID           string
	Email        string
	Name         string
	AvatarURL    string
	AuthProvider string // "google" or "magic_link"
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
