package auth

import (
	"github.com/nishiki/frontend/pkg/api/common"
	"github.com/nishiki/frontend/pkg/types"
)

// Client handles authentication-related API calls
type Client struct {
	common *common.Client
}

// NewClient creates a new auth API client
func NewClient(commonClient *common.Client) *Client {
	return &Client{
		common: commonClient,
	}
}

// GetCurrentUser gets the currently authenticated user
func (c *Client) GetCurrentUser() (*types.User, error) {
	resp, err := c.common.Get("/auth/me")
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.User](resp)
}

// GetOIDCConfig gets the OIDC configuration from the backend
func (c *Client) GetOIDCConfig() (map[string]interface{}, error) {
	resp, err := c.common.Get("/auth/oidc-config")
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[map[string]interface{}](resp)
}
