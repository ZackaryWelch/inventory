package types

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthInfoResponse represents the response from /auth/me
type AuthInfoResponse struct {
	User   User       `json:"user"`
	Claims ClaimsInfo `json:"claims"`
}

// ClaimsInfo represents JWT claims information
type ClaimsInfo struct {
	Subject   string   `json:"subject"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	Groups    []string `json:"groups"`
	ExpiresAt int64    `json:"expires_at"`
	IssuedAt  int64    `json:"issued_at"`
}
