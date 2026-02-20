package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/usecases"
)

func registerTools(s *mcp.Server, mctx *MCPContext) {
	registerCollectionTools(s, mctx)
	registerContainerTools(s, mctx)
	registerObjectTools(s, mctx)
	registerGroupTools(s, mctx)
	registerImportTools(s, mctx)
}

// --- Collection tools ---

func registerCollectionTools(s *mcp.Server, mctx *MCPContext) {
	type CreateCollectionInput struct {
		Name       string  `json:"name" jsonschema:"Name of the collection"`
		ObjectType string  `json:"object_type" jsonschema:"Object type: food, book, videogame, music, boardgame, general"`
		Location   string  `json:"location,omitempty" jsonschema:"Physical location of the collection"`
		GroupID    string  `json:"group_id,omitempty" jsonschema:"Group ID to share this collection with (optional)"`
		Tags       []string `json:"tags,omitempty" jsonschema:"Tags for the collection"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_collection",
		Description: "Create a new inventory collection for a specific object type (food, books, games, etc.)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateCollectionInput) (*mcp.CallToolResult, any, error) {
		var groupID *entities.GroupID
		if input.GroupID != "" {
			gid, err := entities.GroupIDFromString(input.GroupID)
			if err != nil {
				r, _ := errorResult(fmt.Errorf("invalid group_id: %w", err))
				return r, nil, nil
			}
			groupID = &gid
		}

		ucReq := usecases.CreateCollectionRequest{
			UserID:     mctx.userID(),
			GroupID:    groupID,
			Name:       input.Name,
			ObjectType: entities.ObjectType(input.ObjectType),
			Tags:       input.Tags,
			Location:   input.Location,
			UserToken:  mctx.Token,
		}
		resp, err := mctx.createCollectionUC().Execute(ctx, ucReq)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(response.NewCollectionResponse(resp.Collection))
		return r, nil, err
	})

	type UpdateCollectionInput struct {
		CollectionID string   `json:"collection_id" jsonschema:"ID of the collection to update"`
		Name         string   `json:"name,omitempty" jsonschema:"New name for the collection (optional)"`
		Location     string   `json:"location,omitempty" jsonschema:"New location (optional)"`
		Tags         []string `json:"tags,omitempty" jsonschema:"New tags (optional, replaces existing)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_collection",
		Description: "Update a collection's name, location, or tags",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateCollectionInput) (*mcp.CallToolResult, any, error) {
		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.UpdateCollectionRequest{
			CollectionID: collectionID,
			UserID:       mctx.userID(),
			Tags:         input.Tags,
			UserToken:    mctx.Token,
		}
		if input.Name != "" {
			ucReq.Name = &input.Name
		}
		if input.Location != "" {
			ucReq.Location = &input.Location
		}

		resp, err := mctx.updateCollectionUC().Execute(ctx, ucReq)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(response.NewCollectionResponse(resp.Collection))
		return r, nil, err
	})

	type DeleteCollectionInput struct {
		CollectionID string `json:"collection_id" jsonschema:"ID of the collection to delete"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_collection",
		Description: "Delete a collection (must have no containers)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteCollectionInput) (*mcp.CallToolResult, any, error) {
		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		_, err = mctx.deleteCollectionUC().Execute(ctx, usecases.DeleteCollectionRequest{
			CollectionID: collectionID,
			UserID:       mctx.userID(),
		})
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(map[string]any{"success": true, "collection_id": input.CollectionID})
		return r, nil, err
	})
}

// --- Container tools ---

func registerContainerTools(s *mcp.Server, mctx *MCPContext) {
	type CreateContainerInput struct {
		CollectionID      string   `json:"collection_id" jsonschema:"ID of the collection to add this container to"`
		Name              string   `json:"name" jsonschema:"Name of the container"`
		ContainerType     string   `json:"container_type" jsonschema:"Type: room, bookshelf, shelf, binder, cabinet, general"`
		ParentContainerID string   `json:"parent_container_id,omitempty" jsonschema:"ID of parent container (optional)"`
		Location          string   `json:"location,omitempty" jsonschema:"Physical location within the collection"`
		Capacity          *float64 `json:"capacity,omitempty" jsonschema:"Maximum capacity (optional)"`
		Width             *float64 `json:"width,omitempty" jsonschema:"Width dimension (optional)"`
		Depth             *float64 `json:"depth,omitempty" jsonschema:"Depth dimension (optional)"`
		Rows              *int     `json:"rows,omitempty" jsonschema:"Number of rows (optional)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_container",
		Description: "Create a container within a collection for organizing objects",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateContainerInput) (*mcp.CallToolResult, any, error) {
		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.CreateContainerRequest{
			CollectionID:  collectionID,
			Name:          input.Name,
			ContainerType: entities.ContainerType(input.ContainerType),
			Location:      input.Location,
			Capacity:      input.Capacity,
			Width:         input.Width,
			Depth:         input.Depth,
			Rows:          input.Rows,
			UserID:        mctx.userID(),
			UserToken:     mctx.Token,
		}

		if input.ParentContainerID != "" {
			parentID, err := entities.ContainerIDFromString(input.ParentContainerID)
			if err != nil {
				r, _ := errorResult(fmt.Errorf("invalid parent_container_id: %w", err))
				return r, nil, nil
			}
			ucReq.ParentContainerID = &parentID
		}

		resp, err := mctx.createContainerUC().Execute(ctx, ucReq)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(response.NewContainerResponse(resp.Container))
		return r, nil, err
	})

	type UpdateContainerInput struct {
		ContainerID   string   `json:"container_id" jsonschema:"ID of the container to update"`
		Name          string   `json:"name,omitempty" jsonschema:"New name (optional)"`
		ContainerType string   `json:"container_type,omitempty" jsonschema:"New type (optional): room, bookshelf, shelf, binder, cabinet, general"`
		Location      string   `json:"location,omitempty" jsonschema:"New location (optional)"`
		Capacity      *float64 `json:"capacity,omitempty" jsonschema:"New capacity (optional)"`
		Width         *float64 `json:"width,omitempty" jsonschema:"New width (optional)"`
		Depth         *float64 `json:"depth,omitempty" jsonschema:"New depth (optional)"`
		Rows          *int     `json:"rows,omitempty" jsonschema:"New row count (optional)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_container",
		Description: "Update a container's name, type, location, or dimensions",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateContainerInput) (*mcp.CallToolResult, any, error) {
		containerID, err := entities.ContainerIDFromString(input.ContainerID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid container_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.UpdateContainerRequest{
			ContainerID: containerID,
			UserID:      mctx.userID(),
			UserToken:   mctx.Token,
		}
		if input.Name != "" {
			ucReq.Name = &input.Name
		}
		if input.ContainerType != "" {
			ct := entities.ContainerType(input.ContainerType)
			ucReq.ContainerType = &ct
		}
		if input.Location != "" {
			ucReq.Location = &input.Location
		}
		if input.Capacity != nil {
			ucReq.Capacity = &input.Capacity
		}
		if input.Width != nil {
			ucReq.Width = &input.Width
		}
		if input.Depth != nil {
			ucReq.Depth = &input.Depth
		}
		if input.Rows != nil {
			ucReq.Rows = &input.Rows
		}

		resp, err := mctx.updateContainerUC().Execute(ctx, ucReq)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(response.NewContainerResponse(resp.Container))
		return r, nil, err
	})

	type DeleteContainerInput struct {
		ContainerID string `json:"container_id" jsonschema:"ID of the container to delete"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_container",
		Description: "Delete a container (must have no child containers)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteContainerInput) (*mcp.CallToolResult, any, error) {
		containerID, err := entities.ContainerIDFromString(input.ContainerID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid container_id: %w", err))
			return r, nil, nil
		}

		_, err = mctx.deleteContainerUC().Execute(ctx, usecases.DeleteContainerRequest{
			ContainerID: containerID,
			UserID:      mctx.userID(),
			UserToken:   mctx.Token,
		})
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(map[string]any{"success": true, "container_id": input.ContainerID})
		return r, nil, err
	})
}

// --- Object tools ---

func registerObjectTools(s *mcp.Server, mctx *MCPContext) {
	type CreateObjectInput struct {
		ContainerID string                 `json:"container_id" jsonschema:"ID of the container to add the object to"`
		Name        string                 `json:"name" jsonschema:"Name of the object"`
		Description string                 `json:"description,omitempty" jsonschema:"Description (optional)"`
		ObjectType  string                 `json:"object_type" jsonschema:"Object type matching the collection: food, book, videogame, music, boardgame, general"`
		Quantity    *float64               `json:"quantity,omitempty" jsonschema:"Quantity (optional)"`
		Unit        string                 `json:"unit,omitempty" jsonschema:"Unit of quantity e.g. kg, pieces (optional)"`
		Properties  map[string]interface{} `json:"properties,omitempty" jsonschema:"Type-specific properties e.g. author, ISBN, brand (optional)"`
		Tags        []string               `json:"tags,omitempty" jsonschema:"Tags (optional)"`
		ExpiresAt   string                 `json:"expires_at,omitempty" jsonschema:"Expiration date in RFC3339 format (optional, mainly for food)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_object",
		Description: "Add an object to a container in a collection",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateObjectInput) (*mcp.CallToolResult, any, error) {
		containerID, err := entities.ContainerIDFromString(input.ContainerID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid container_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.CreateObjectRequest{
			ContainerID: containerID,
			Name:        input.Name,
			Description: input.Description,
			ObjectType:  entities.ObjectType(input.ObjectType),
			Quantity:    input.Quantity,
			Unit:        input.Unit,
			Properties:  input.Properties,
			Tags:        input.Tags,
			UserID:      mctx.userID(),
			UserToken:   mctx.Token,
		}

		if input.ExpiresAt != "" {
			t, err := time.Parse(time.RFC3339, input.ExpiresAt)
			if err != nil {
				r, _ := errorResult(fmt.Errorf("invalid expires_at format (use RFC3339): %w", err))
				return r, nil, nil
			}
			ucReq.ExpiresAt = &t
		}

		resp, err := mctx.createObjectUC().Execute(ctx, ucReq)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(response.NewObjectResponse(*resp.Object))
		return r, nil, err
	})

	type DeleteObjectInput struct {
		ContainerID string `json:"container_id" jsonschema:"ID of the container that holds the object"`
		ObjectID    string `json:"object_id" jsonschema:"ID of the object to delete"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_object",
		Description: "Remove an object from a container",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteObjectInput) (*mcp.CallToolResult, any, error) {
		containerID, err := entities.ContainerIDFromString(input.ContainerID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid container_id: %w", err))
			return r, nil, nil
		}
		objectID, err := entities.ObjectIDFromHex(input.ObjectID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid object_id: %w", err))
			return r, nil, nil
		}

		_, err = mctx.deleteObjectUC().Execute(ctx, usecases.DeleteObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      mctx.userID(),
			UserToken:   mctx.Token,
		})
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(map[string]any{"success": true, "object_id": input.ObjectID})
		return r, nil, err
	})

	type UpdateObjectInput struct {
		ContainerID string                 `json:"container_id" jsonschema:"ID of the container holding the object"`
		ObjectID    string                 `json:"object_id" jsonschema:"ID of the object to update"`
		Name        string                 `json:"name,omitempty" jsonschema:"New name (optional)"`
		Properties  map[string]interface{} `json:"properties,omitempty" jsonschema:"New properties (optional, replaces existing)"`
		Tags        []string               `json:"tags,omitempty" jsonschema:"New tags (optional, replaces existing)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_object",
		Description: "Update an object's name, properties, or tags",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateObjectInput) (*mcp.CallToolResult, any, error) {
		containerID, err := entities.ContainerIDFromString(input.ContainerID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid container_id: %w", err))
			return r, nil, nil
		}
		objectID, err := entities.ObjectIDFromHex(input.ObjectID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid object_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.UpdateObjectRequest{
			ContainerID: containerID,
			ObjectID:    objectID,
			UserID:      mctx.userID(),
			UserToken:   mctx.Token,
		}
		if input.Name != "" {
			ucReq.Name = &input.Name
		}
		if input.Properties != nil {
			ucReq.Properties = input.Properties
		}
		if input.Tags != nil {
			ucReq.Tags = input.Tags
		}

		resp, err := mctx.updateObjectUC().Execute(ctx, ucReq)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(response.NewObjectResponse(*resp.Object))
		return r, nil, err
	})
}

// --- Group tools ---

func registerGroupTools(s *mcp.Server, mctx *MCPContext) {
	type CreateGroupInput struct {
		Name string `json:"name" jsonschema:"Name of the new group"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_group",
		Description: "Create a new group for collaborating on collections",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateGroupInput) (*mcp.CallToolResult, any, error) {
		resp, err := mctx.createGroupUC().Execute(ctx, usecases.CreateGroupRequest{
			Name:      input.Name,
			CreatorID: mctx.userID(),
			UserToken: mctx.Token,
		})
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(response.NewGroupResponse(resp.Group))
		return r, nil, err
	})

	// Stubs for unimplemented group operations.
	type JoinGroupInput struct {
		InviteCode string `json:"invite_code" jsonschema:"Invitation code for the group"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "join_group",
		Description: "Join a group using an invite code (currently unavailable — backend returns 501)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input JoinGroupInput) (*mcp.CallToolResult, any, error) {
		r, _ := errorResult(fmt.Errorf("backend unimplemented: JoinGroup returns 501. Fix planned in Phase 2"))
		return r, nil, nil
	})

	type UpdateGroupInput struct {
		GroupID string `json:"group_id" jsonschema:"ID of the group to update"`
		Name    string `json:"name" jsonschema:"New name for the group"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_group",
		Description: "Update a group's name (currently unavailable — no backend endpoint exists)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateGroupInput) (*mcp.CallToolResult, any, error) {
		r, _ := errorResult(fmt.Errorf("backend missing: no update endpoint for groups. Fix planned in Phase 2"))
		return r, nil, nil
	})

	type DeleteGroupInput struct {
		GroupID string `json:"group_id" jsonschema:"ID of the group to delete"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_group",
		Description: "Delete a group (currently unavailable — no backend endpoint exists)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteGroupInput) (*mcp.CallToolResult, any, error) {
		r, _ := errorResult(fmt.Errorf("backend missing: no delete endpoint for groups. Fix planned in Phase 2"))
		return r, nil, nil
	})
}

// --- Import tools ---

func registerImportTools(s *mcp.Server, mctx *MCPContext) {
	type BulkImportInput struct {
		CollectionID     string                   `json:"collection_id" jsonschema:"ID of the collection to import into"`
		Data             []map[string]interface{} `json:"data" jsonschema:"Array of objects to import, each must have a 'name' field"`
		DistributionMode string                   `json:"distribution_mode,omitempty" jsonschema:"How to distribute objects: automatic, target, or manual (default)"`
		TargetContainerID string                  `json:"target_container_id,omitempty" jsonschema:"Container ID for target distribution mode (optional)"`
		DefaultTags      []string                 `json:"default_tags,omitempty" jsonschema:"Tags to apply to all imported objects (optional)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "bulk_import",
		Description: "Bulk import objects into a collection. Each item must have a 'name' field; other fields become properties.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input BulkImportInput) (*mcp.CallToolResult, any, error) {
		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.BulkImportCollectionRequest{
			UserID:           mctx.userID(),
			CollectionID:     collectionID,
			DistributionMode: input.DistributionMode,
			Data:             input.Data,
			DefaultTags:      input.DefaultTags,
			UserToken:        mctx.Token,
		}

		if input.TargetContainerID != "" {
			targetID, err := entities.ContainerIDFromString(input.TargetContainerID)
			if err != nil {
				r, _ := errorResult(fmt.Errorf("invalid target_container_id: %w", err))
				return r, nil, nil
			}
			ucReq.TargetContainerID = &targetID
		}

		resp, err := mctx.bulkImportCollectionUC().Execute(ctx, ucReq)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(resp)
		return r, nil, err
	})
}
