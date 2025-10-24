package auth

import (
	"github.com/nishiki/frontend/pkg/api/common"
	"github.com/nishiki/frontend/pkg/types"
)

// Client handles authentication-related API calls
type Client struct {
	common   *common.Client
	clientID string
}

// NewClient creates a new auth API client
func NewClient(commonClient *common.Client, clientID string) *Client {
	return &Client{
		common:   commonClient,
		clientID: clientID,
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
func (c *Client) GetOIDCConfig() (*map[string]interface{}, error) {
	// Add client_id query parameter as required by backend
	endpoint := "/auth/oidc-config?client_id=" + c.clientID
	resp, err := c.common.Get(endpoint)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[map[string]interface{}](resp)
}
