package collections

import (
	"fmt"

	"github.com/nishiki/frontend/pkg/api/common"
	"github.com/nishiki/frontend/pkg/types"
)

// Client handles collection-related API calls
type Client struct {
	common *common.Client
}

// NewClient creates a new collections API client
func NewClient(commonClient *common.Client) *Client {
	return &Client{
		common: commonClient,
	}
}

// List gets all collections for a user
func (c *Client) List(accountID string) ([]types.Collection, error) {
	resp, err := c.common.Get(fmt.Sprintf("/accounts/%s/collections", accountID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponseList[types.Collection](resp)
}

// Get gets a specific collection by ID
func (c *Client) Get(accountID, collectionID string) (*types.Collection, error) {
	resp, err := c.common.Get(fmt.Sprintf("/accounts/%s/collections/%s", accountID, collectionID))
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Collection](resp)
}

// Create creates a new collection
func (c *Client) Create(accountID string, req types.CreateCollectionRequest) (*types.Collection, error) {
	resp, err := c.common.Post(fmt.Sprintf("/accounts/%s/collections", accountID), req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Collection](resp)
}

// Update updates an existing collection
func (c *Client) Update(accountID, collectionID string, req types.UpdateCollectionRequest) (*types.Collection, error) {
	resp, err := c.common.Put(fmt.Sprintf("/accounts/%s/collections/%s", accountID, collectionID), req)
	if err != nil {
		return nil, err
	}

	return common.DecodeResponse[types.Collection](resp)
}

// Delete deletes a collection. If force is true, cascade-deletes containers and objects.
func (c *Client) Delete(accountID, collectionID string, force bool) error {
	path := fmt.Sprintf("/accounts/%s/collections/%s", accountID, collectionID)
	if force {
		path += "?force=true"
	}
	resp, err := c.common.Delete(path)
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}

// UpdateSchema updates the property schema for a collection
func (c *Client) UpdateSchema(accountID, collectionID string, req types.UpdatePropertySchemaRequest) error {
	resp, err := c.common.Put(fmt.Sprintf("/accounts/%s/collections/%s/schema", accountID, collectionID), req)
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}

// ImportObjects imports objects to a collection in bulk
func (c *Client) ImportObjects(accountID, collectionID string, req types.BulkImportCollectionRequest) error {
	resp, err := c.common.Post(fmt.Sprintf("/accounts/%s/collections/%s/import", accountID, collectionID), req)
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}
