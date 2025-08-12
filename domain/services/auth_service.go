//go:generate mockgen -source=auth_service.go -destination=../../mocks/mock_auth_service.go -package=mocks

package services

import (
	"context"

	"github.com/nishiki/backend-go/domain/entities"
)

type AuthClaims struct {
	Subject   string   `json:"sub"`
	Email     string   `json:"email"`
	Username  string   `json:"preferred_username"`
	Groups    []string `json:"groups"`
	Name      string   `json:"name"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
	Issuer    string   `json:"iss"`
	Audience  string   `json:"aud"`
}

type AuthService interface {
	ValidateToken(ctx context.Context, token string) (*AuthClaims, error)
	GetUserFromClaims(ctx context.Context, claims *AuthClaims) (*entities.User, error)
	CreateUserFromClaims(ctx context.Context, claims *AuthClaims) (*entities.User, error)

	// OIDC proxy methods for frontend integration
	GetOIDCConfig(ctx context.Context) (map[string]interface{}, error)
	ProxyTokenExchange(ctx context.Context, tokenRequest map[string]interface{}) ([]byte, int, error)

	// Group and user fetching from Authentik (now requires user's JWT token)
	CreateGroup(ctx context.Context, userToken, name string, creatorID string) (*entities.Group, error)
	GetUserGroups(ctx context.Context, userToken, userID string) ([]*entities.Group, error)
	GetGroupUsers(ctx context.Context, userToken, groupID string) ([]*entities.User, error)
	GetUserByID(ctx context.Context, userToken, userID string) (*entities.User, error)
	GetGroupByID(ctx context.Context, userToken, groupID string) (*entities.Group, error)
}
