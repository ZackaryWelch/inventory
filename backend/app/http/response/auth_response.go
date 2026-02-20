package response

import (
	"time"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/services"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserListResponse []UserResponse

type AuthInfoResponse struct {
	User   UserResponse `json:"user"`
	Claims ClaimsInfo   `json:"claims"`
}

type ClaimsInfo struct {
	Subject   string   `json:"subject"`
	Email     string   `json:"email"`
	Username  string   `json:"username"`
	Groups    []string `json:"groups"`
	ExpiresAt int64    `json:"expires_at"`
	IssuedAt  int64    `json:"issued_at"`
}

func NewUserResponse(user *entities.User) UserResponse {
	return UserResponse{
		ID:        user.ID().String(),
		Name:      user.Username().String(), // Using username as name
		Email:     user.EmailAddress().String(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}
}

func NewUserListResponse(users []*entities.User) UserListResponse {
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = NewUserResponse(user)
	}

	return UserListResponse(userResponses)
}

func NewAuthInfoResponse(user *entities.User, claims *services.AuthClaims) AuthInfoResponse {
	return AuthInfoResponse{
		User: NewUserResponse(user),
		Claims: ClaimsInfo{
			Subject:   claims.Subject,
			Email:     claims.Email,
			Username:  claims.Username,
			Groups:    claims.Groups,
			ExpiresAt: claims.ExpiresAt,
			IssuedAt:  claims.IssuedAt,
		},
	}
}
