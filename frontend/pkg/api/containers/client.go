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

// List gets all containers (optionally filtered by collection_id)
func (c *Client) List(accountID, collectionID string) ([]types.Container, error) {
	// Backend uses /containers with optional collection_id query param
	resp, err := c.common.Get(fmt.Sprintf("/containers?collection_id=%s", collectionID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.Container](resp)
}

// Get gets a specific container by ID
func (c *Client) Get(accountID, collectionID, containerID string) (*types.Container, error) {
	resp, err := c.common.Get(fmt.Sprintf("/containers/%s", containerID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Container](resp)
}

// GetChildren gets all child containers of a parent container
func (c *Client) GetChildren(accountID, collectionID, parentContainerID string) ([]types.Container, error) {
	// Filter containers by parent_container_id
	resp, err := c.common.Get(fmt.Sprintf("/containers?collection_id=%s&parent_id=%s", collectionID, parentContainerID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.Container](resp)
}

// Create creates a new container
func (c *Client) Create(accountID, collectionID string, req types.CreateContainerRequest) (*types.Container, error) {
	// Backend uses /containers (collection_id is in request body)
	resp, err := c.common.Post("/containers", req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Container](resp)
}

// Update updates an existing container
func (c *Client) Update(accountID, collectionID, containerID string, req types.UpdateContainerRequest) (*types.Container, error) {
	resp, err := c.common.Put(fmt.Sprintf("/containers/%s", containerID), req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Container](resp)
}

// Delete deletes a container
func (c *Client) Delete(accountID, collectionID, containerID string) error {
	resp, err := c.common.Delete(fmt.Sprintf("/containers/%s", containerID))
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}

// GetObjects gets all objects in a container
func (c *Client) GetObjects(accountID, collectionID, containerID string) ([]types.Object, error) {
	resp, err := c.common.Get(fmt.Sprintf("/containers/%s/objects", containerID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.Object](resp)
}
