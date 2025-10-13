package categories

import (
	"fmt"

	"github.com/nishiki/frontend/pkg/api/common"
	"github.com/nishiki/frontend/pkg/types"
)

// Client handles category-related API calls
type Client struct {
	common *common.Client
}

// NewClient creates a new categories API client
func NewClient(commonClient *common.Client) *Client {
	return &Client{
		common: commonClient,
	}
}

// List gets all categories
func (c *Client) List() ([]types.Category, error) {
	resp, err := c.common.Get("/categories")
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.Category](resp)
}

// Get gets a specific category by ID
func (c *Client) Get(id string) (*types.Category, error) {
	resp, err := c.common.Get("/categories/" + id)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Category](resp)
}

// Create creates a new category
func (c *Client) Create(req types.CreateCategoryRequest) (*types.Category, error) {
	resp, err := c.common.Post("/categories", req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Category](resp)
}

// Update updates an existing category
func (c *Client) Update(id string, req types.UpdateCategoryRequest) (*types.Category, error) {
	resp, err := c.common.Put(fmt.Sprintf("/categories/%s", id), req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Category](resp)
}

// Delete deletes a category
func (c *Client) Delete(id string) error {
	resp, err := c.common.Delete("/categories/" + id)
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}
