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

	// Backend returns wrapped response: {collections: [...], total: N}
	listResp, err := common.DecodeResponse[types.CollectionListResponse](resp)
	if err != nil {
		return nil, err
	}

	return listResp.Collections, nil
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

// Delete deletes a collection
func (c *Client) Delete(accountID, collectionID string) error {
	resp, err := c.common.Delete(fmt.Sprintf("/accounts/%s/collections/%s", accountID, collectionID))
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}

// ImportObjects imports objects to a collection in bulk
func (c *Client) ImportObjects(accountID, collectionID string, req types.BulkImportRequest) error {
	resp, err := c.common.Post(fmt.Sprintf("/accounts/%s/collections/%s/import", accountID, collectionID), req)
	if err != nil {
		return err
	}

	return common.CheckResponse(resp)
}
