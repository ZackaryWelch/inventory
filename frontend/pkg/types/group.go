package types

import "time"

// Group represents a group in the system
type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Members     []User    `json:"members,omitempty"`
}

// CreateGroupRequest represents the request to create a new group
type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateGroupRequest represents the request to update a group
type UpdateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// InviteUserRequest represents the request to invite a user to a group
type InviteUserRequest struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}
