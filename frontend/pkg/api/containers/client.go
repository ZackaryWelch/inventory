package containers

import (
	"fmt"

	"github.com/nishiki/frontend/pkg/api/common"
	"github.com/nishiki/frontend/pkg/types"
)

// Client handles container-related API calls
type Client struct {
	common *common.Client
}

// NewClient creates a new containers API client
func NewClient(commonClient *common.Client) *Client {
	return &Client{
		common: commonClient,
	}
}

// List gets all containers for a collection
func (c *Client) List(accountID, collectionID string) ([]types.Container, error) {
	resp, err := c.common.Get(fmt.Sprintf("/accounts/%s/collections/%s/containers", accountID, collectionID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.Container](resp)
}

// Get gets a specific container by ID
func (c *Client) Get(accountID, collectionID, containerID string) (*types.Container, error) {
	resp, err := c.common.Get(fmt.Sprintf("/accounts/%s/collections/%s/containers/%s", accountID, collectionID, containerID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Container](resp)
}

// Create creates a new container
func (c *Client) Create(accountID, collectionID string, req types.CreateContainerRequest) (*types.Container, error) {
	resp, err := c.common.Post(fmt.Sprintf("/accounts/%s/collections/%s/containers", accountID, collectionID), req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Container](resp)
}

// Update updates an existing container
func (c *Client) Update(accountID, collectionID, containerID string, req types.UpdateContainerRequest) (*types.Container, error) {
	resp, err := c.common.Put(fmt.Sprintf("/accounts/%s/collections/%s/containers/%s", accountID, collectionID, containerID), req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Container](resp)
}

// Delete deletes a container
func (c *Client) Delete(accountID, collectionID, containerID string) error {
	resp, err := c.common.Delete(fmt.Sprintf("/accounts/%s/collections/%s/containers/%s", accountID, collectionID, containerID))
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}

// GetObjects gets all objects in a container
func (c *Client) GetObjects(accountID, collectionID, containerID string) ([]types.Object, error) {
	resp, err := c.common.Get(fmt.Sprintf("/accounts/%s/collections/%s/containers/%s/objects", accountID, collectionID, containerID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.Object](resp)
}
