package objects

import (
	"fmt"

	"github.com/nishiki/frontend/pkg/api/common"
	"github.com/nishiki/frontend/pkg/types"
)

// Client handles object-related API calls
type Client struct {
	common *common.Client
}

// NewClient creates a new objects API client
func NewClient(commonClient *common.Client) *Client {
	return &Client{
		common: commonClient,
	}
}

// Get gets a specific object by ID
func (c *Client) Get(accountID, objectID string) (*types.Object, error) {
	resp, err := c.common.Get(fmt.Sprintf("/accounts/%s/objects/%s", accountID, objectID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Object](resp)
}

// Create creates a new object
func (c *Client) Create(accountID string, req types.CreateObjectRequest) (*types.Object, error) {
	resp, err := c.common.Post(fmt.Sprintf("/accounts/%s/objects", accountID), req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Object](resp)
}

// Update updates an existing object
func (c *Client) Update(accountID, objectID string, req types.UpdateObjectRequest) (*types.Object, error) {
	resp, err := c.common.Put(fmt.Sprintf("/accounts/%s/objects/%s", accountID, objectID), req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Object](resp)
}

// Delete deletes an object
func (c *Client) Delete(accountID, objectID, containerID string) error {
	url := fmt.Sprintf("/accounts/%s/objects/%s?container_id=%s", accountID, objectID, containerID)
	resp, err := c.common.Delete(url)
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}

// Search searches for objects based on filter criteria
func (c *Client) Search(accountID string, filter types.SearchFilter) ([]types.SearchResult, error) {
	// Convert filter to query parameters or POST body as needed
	resp, err := c.common.Post(fmt.Sprintf("/accounts/%s/objects/search", accountID), filter)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.SearchResult](resp)
}

// Move moves an object to a different container
func (c *Client) Move(accountID, objectID, newContainerID string) (*types.Object, error) {
	req := map[string]string{"container_id": newContainerID}
	resp, err := c.common.Put(fmt.Sprintf("/accounts/%s/objects/%s/move", accountID, objectID), req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Object](resp)
}
