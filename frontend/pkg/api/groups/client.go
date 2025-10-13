package groups

import (
	"fmt"

	"github.com/nishiki/frontend/pkg/api/common"
	"github.com/nishiki/frontend/pkg/types"
)

// Client handles group-related API calls
type Client struct {
	common *common.Client
}

// NewClient creates a new groups API client
func NewClient(commonClient *common.Client) *Client {
	return &Client{
		common: commonClient,
	}
}

// List gets all groups for the current user
func (c *Client) List() ([]types.Group, error) {
	resp, err := c.common.Get("/groups")
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.Group](resp)
}

// Get gets a specific group by ID
func (c *Client) Get(id string) (*types.Group, error) {
	resp, err := c.common.Get("/groups/" + id)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Group](resp)
}

// Create creates a new group
func (c *Client) Create(req types.CreateGroupRequest) (*types.Group, error) {
	resp, err := c.common.Post("/groups", req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Group](resp)
}

// Update updates an existing group
func (c *Client) Update(id string, req types.UpdateGroupRequest) (*types.Group, error) {
	resp, err := c.common.Put("/groups/"+id, req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Group](resp)
}

// Delete deletes a group
func (c *Client) Delete(id string) error {
	resp, err := c.common.Delete("/groups/" + id)
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}

// GetMembers gets all members of a group
func (c *Client) GetMembers(id string) ([]types.User, error) {
	resp, err := c.common.Get("/groups/" + id + "/users")
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.User](resp)
}

// InviteUser invites a user to join a group
func (c *Client) InviteUser(groupID string, req types.InviteUserRequest) error {
	resp, err := c.common.Post(fmt.Sprintf("/groups/%s/invite", groupID), req)
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}

// RemoveMember removes a member from a group
func (c *Client) RemoveMember(groupID, userID string) error {
	resp, err := c.common.Delete(fmt.Sprintf("/groups/%s/members/%s", groupID, userID))
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}

// JoinByHash joins a group using an invite hash
func (c *Client) JoinByHash(inviteHash string) (*types.Group, error) {
	req := map[string]string{"invite_hash": inviteHash}
	resp, err := c.common.Post("/groups/join", req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Group](resp)
}
