package mcpserver

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/nishiki/backend/app/http/response"
	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/usecases"
)

func registerTools(s *mcp.Server, mctx *MCPContext) {
	registerCollectionTools(s, mctx)
	registerContainerTools(s, mctx)
	registerObjectTools(s, mctx)
	registerGroupTools(s, mctx)
	registerImportTools(s, mctx)
	registerSchemaTools(s, mctx)
	registerExportTools(s, mctx)
	registerSearchTools(s, mctx)
}

// --- Collection tools ---

func registerCollectionTools(s *mcp.Server, mctx *MCPContext) {
	type CreateCollectionInput struct {
		Name       string   `json:"name" jsonschema:"Name of the collection"`
		ObjectType string   `json:"object_type" jsonschema:"Object type: food, book, videogame, music, boardgame, general"`
		Location   string   `json:"location,omitempty" jsonschema:"Physical location of the collection"`
		GroupID    string   `json:"group_id,omitempty" jsonschema:"Group ID to share this collection with (optional)"`
		Tags       []string `json:"tags,omitempty" jsonschema:"Tags for the collection"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "create_collection",
		Description: "Create a new inventory collection for a specific object type (food, books, games, etc.)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input CreateCollectionInput) (*mcp.CallToolResult, any, error) {
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

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
			UserID:     user.ID(),
			GroupID:    groupID,
			Name:       input.Name,
			ObjectType: entities.ObjectType(input.ObjectType),
			Tags:       input.Tags,
			Location:   input.Location,
			UserToken:  token,
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.UpdateCollectionRequest{
			CollectionID: collectionID,
			UserID:       user.ID(),
			Tags:         input.Tags,
			UserToken:    token,
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
		user, _, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		_, err = mctx.deleteCollectionUC().Execute(ctx, usecases.DeleteCollectionRequest{
			CollectionID: collectionID,
			UserID:       user.ID(),
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

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
			UserID:        user.ID(),
			UserToken:     token,
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		containerID, err := entities.ContainerIDFromString(input.ContainerID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid container_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.UpdateContainerRequest{
			ContainerID: containerID,
			UserID:      user.ID(),
			UserToken:   token,
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		containerID, err := entities.ContainerIDFromString(input.ContainerID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid container_id: %w", err))
			return r, nil, nil
		}

		_, err = mctx.deleteContainerUC().Execute(ctx, usecases.DeleteContainerRequest{
			ContainerID: containerID,
			UserID:      user.ID(),
			UserToken:   token,
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

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
			UserID:      user.ID(),
			UserToken:   token,
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
		r, err := jsonResult(response.NewObjectResponse(*resp.Object, input.ContainerID))
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

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
			ContainerID: &containerID,
			ObjectID:    objectID,
			UserID:      user.ID(),
			UserToken:   token,
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

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
			UserID:      user.ID(),
			UserToken:   token,
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
		r, err := jsonResult(response.NewObjectResponse(*resp.Object, input.ContainerID))
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
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		resp, err := mctx.createGroupUC().Execute(ctx, usecases.CreateGroupRequest{
			Name:      input.Name,
			CreatorID: user.ID(),
			UserToken: token,
		})
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(response.NewGroupResponse(resp.Group))
		return r, nil, err
	})

	type AddGroupMemberInput struct {
		GroupID string `json:"group_id" jsonschema:"ID of the group"`
		UserID  string `json:"user_id" jsonschema:"Numeric user ID to add"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "add_group_member",
		Description: "Add a user to a group by their numeric user ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input AddGroupMemberInput) (*mcp.CallToolResult, any, error) {
		_, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		groupID, err := entities.GroupIDFromString(input.GroupID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid group_id: %w", err))
			return r, nil, nil
		}

		if err := mctx.groupUC().AddMember(ctx, usecases.GroupMemberRequest{
			GroupID:   groupID,
			UserID:    input.UserID,
			UserToken: token,
		}); err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(map[string]string{"status": "added"})
		return r, nil, err
	})

	type RemoveGroupMemberInput struct {
		GroupID string `json:"group_id" jsonschema:"ID of the group"`
		UserID  string `json:"user_id" jsonschema:"Numeric user ID to remove"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "remove_group_member",
		Description: "Remove a user from a group by their numeric user ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input RemoveGroupMemberInput) (*mcp.CallToolResult, any, error) {
		_, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		groupID, err := entities.GroupIDFromString(input.GroupID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid group_id: %w", err))
			return r, nil, nil
		}

		if err := mctx.groupUC().RemoveMember(ctx, usecases.GroupMemberRequest{
			GroupID:   groupID,
			UserID:    input.UserID,
			UserToken: token,
		}); err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(map[string]string{"status": "removed"})
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
		Description: "Rename a group.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateGroupInput) (*mcp.CallToolResult, any, error) {
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		groupID, err := entities.GroupIDFromString(input.GroupID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid group_id: %w", err))
			return r, nil, nil
		}

		resp, err := mctx.groupUC().UpdateGroup(ctx, usecases.UpdateGroupRequest{
			GroupID:   groupID,
			Name:      input.Name,
			UserToken: token,
		})
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		_ = user
		r, err := jsonResult(response.NewGroupResponse(resp.Group))
		return r, nil, err
	})

	type DeleteGroupInput struct {
		GroupID string `json:"group_id" jsonschema:"ID of the group to delete"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "delete_group",
		Description: "Delete a group by ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DeleteGroupInput) (*mcp.CallToolResult, any, error) {
		_, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		groupID, err := entities.GroupIDFromString(input.GroupID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid group_id: %w", err))
			return r, nil, nil
		}

		if err := mctx.groupUC().DeleteGroup(ctx, usecases.DeleteGroupRequest{
			GroupID:   groupID,
			UserToken: token,
		}); err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(map[string]string{"status": "deleted"})
		return r, nil, err
	})
}

// --- Import tools ---

func registerImportTools(s *mcp.Server, mctx *MCPContext) {
	type BulkImportInput struct {
		CollectionID      string                   `json:"collection_id" jsonschema:"ID of the collection to import into"`
		Data              []map[string]interface{} `json:"data" jsonschema:"Array of objects to import, each must have a 'name' field"`
		DistributionMode  string                   `json:"distribution_mode,omitempty" jsonschema:"How to distribute objects: automatic, location, target, or manual (default)"`
		TargetContainerID string                   `json:"target_container_id,omitempty" jsonschema:"Container ID for target distribution mode (optional)"`
		DefaultTags       []string                 `json:"default_tags,omitempty" jsonschema:"Tags to apply to all imported objects (optional)"`
		LocationColumn    string                   `json:"location_column,omitempty" jsonschema:"Column name used for container mapping in 'location' mode (default: 'location')"`
		NameColumn        string                   `json:"name_column,omitempty" jsonschema:"Column name override for object name (optional, auto-detected by default)"`
		InferSchema       bool                     `json:"infer_schema,omitempty" jsonschema:"Run type inference and save schema to collection (optional)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "bulk_import",
		Description: "Bulk import objects into a collection. Each item must have a 'name' field; other fields become properties. Use distribution_mode='location' to auto-create containers from a Location column.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input BulkImportInput) (*mcp.CallToolResult, any, error) {
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.BulkImportCollectionRequest{
			UserID:           user.ID(),
			CollectionID:     collectionID,
			DistributionMode: input.DistributionMode,
			Data:             input.Data,
			DefaultTags:      input.DefaultTags,
			UserToken:        token,
			LocationColumn:   input.LocationColumn,
			NameColumn:       input.NameColumn,
			InferSchema:      input.InferSchema,
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

	// smart_import: CSV string → parse → type inference → location distribution
	type SmartImportInput struct {
		CollectionID   string   `json:"collection_id" jsonschema:"ID of the collection to import into"`
		CSVData        string   `json:"csv_data" jsonschema:"Raw CSV string content including header row"`
		LocationColumn string   `json:"location_column,omitempty" jsonschema:"Column name for container mapping (default: 'location')"`
		NameColumn     string   `json:"name_column,omitempty" jsonschema:"Column name for object name (optional, auto-detected)"`
		ObjectType     string   `json:"object_type,omitempty" jsonschema:"Object type override (optional, defaults to collection type)"`
		DefaultTags    []string `json:"default_tags,omitempty" jsonschema:"Tags to apply to all imported objects (optional)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "smart_import",
		Description: "Parse a raw CSV string, infer property types, sanitize values, auto-create containers from Location column, and import into a collection. Returns the import summary and inferred schema.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SmartImportInput) (*mcp.CallToolResult, any, error) {
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		// Parse CSV
		data, headers, parseErr := parseCSVString(input.CSVData)
		if parseErr != nil {
			r, _ := errorResult(fmt.Errorf("CSV parse error: %w", parseErr))
			return r, nil, nil
		}
		if len(data) == 0 {
			r, _ := errorResult(fmt.Errorf("CSV contains no data rows"))
			return r, nil, nil
		}

		// Auto-detect location column if not specified
		locationCol := input.LocationColumn
		if locationCol == "" {
			for _, h := range headers {
				if strings.EqualFold(h, "location") {
					locationCol = h
					break
				}
			}
		}

		distMode := "location"
		if locationCol == "" {
			distMode = "" // fall back to default distribution
		}

		ucReq := usecases.BulkImportCollectionRequest{
			UserID:           user.ID(),
			CollectionID:     collectionID,
			DistributionMode: distMode,
			Data:             data,
			DefaultTags:      input.DefaultTags,
			UserToken:        token,
			LocationColumn:   locationCol,
			NameColumn:       input.NameColumn,
			InferSchema:      true,
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

// --- Search tools ---

func registerSearchTools(s *mcp.Server, mctx *MCPContext) {
	type SearchObjectsInput struct {
		CollectionID    string            `json:"collection_id" jsonschema:"ID of the collection to search"`
		Query           string            `json:"query,omitempty" jsonschema:"Case-insensitive substring match on object name (optional)"`
		Tags            []string          `json:"tags,omitempty" jsonschema:"All listed tags must be present on matching objects (optional)"`
		ContainerID     string            `json:"container_id,omitempty" jsonschema:"Restrict search to this container ID (optional)"`
		PropertyFilters map[string]string `json:"property_filters,omitempty" jsonschema:"Key/value pairs: object property must contain the value (case-insensitive, optional)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_objects",
		Description: "Search and filter objects in a collection by name, tags, container, or property values. All filters are ANDed together.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SearchObjectsInput) (*mcp.CallToolResult, any, error) {
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		ucReq := usecases.GetCollectionObjectsRequest{
			CollectionID:    collectionID,
			UserID:          user.ID(),
			UserToken:       token,
			Query:           input.Query,
			Tags:            input.Tags,
			PropertyFilters: input.PropertyFilters,
		}

		if input.ContainerID != "" {
			cid, err := entities.ContainerIDFromString(input.ContainerID)
			if err != nil {
				r, _ := errorResult(fmt.Errorf("invalid container_id: %w", err))
				return r, nil, nil
			}
			ucReq.ContainerID = &cid
		}

		resp, err := mctx.getCollectionObjectsUC().Execute(ctx, ucReq)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		objects := make([]any, len(resp.Objects))
		for i, item := range resp.Objects {
			objects[i] = response.NewObjectResponse(item.Object, item.ContainerID.String())
		}
		r, err := jsonResult(map[string]any{
			"count":   len(objects),
			"objects": objects,
		})
		return r, nil, err
	})
}

// --- Export tools ---

func registerExportTools(s *mcp.Server, mctx *MCPContext) {
	type ExportCollectionInput struct {
		CollectionID string `json:"collection_id" jsonschema:"ID of the collection to export"`
		Format       string `json:"format,omitempty" jsonschema:"Export format: csv or json (default: csv)"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "export_collection",
		Description: "Export all objects in a collection as CSV or JSON. CSV columns follow the collection's property schema order. Useful for data pipelines and backups.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ExportCollectionInput) (*mcp.CallToolResult, any, error) {
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		format := strings.ToLower(strings.TrimSpace(input.Format))
		if format == "" {
			format = "csv"
		}
		if format != "csv" && format != "json" {
			r, _ := errorResult(fmt.Errorf("unsupported format %q: must be csv or json", format))
			return r, nil, nil
		}

		if format == "json" {
			resp, err := mctx.getCollectionObjectsUC().Execute(ctx, usecases.GetCollectionObjectsRequest{
				CollectionID: collectionID,
				UserID:       user.ID(),
				UserToken:    token,
			})
			if err != nil {
				r, _ := errorResult(err)
				return r, nil, nil
			}
			objects := make([]any, len(resp.Objects))
			for i, item := range resp.Objects {
				objects[i] = response.NewObjectResponse(item.Object, item.ContainerID.String())
			}
			r, err := jsonResult(objects)
			return r, nil, err
		}

		// CSV
		resp, err := mctx.exportCollectionUC().Execute(ctx, usecases.ExportCollectionRequest{
			CollectionID: collectionID,
			UserID:       user.ID(),
			UserToken:    token,
		})
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		return textResult(string(resp.CSV)), nil, nil
	})
}

// parseCSVString parses a raw CSV string into rows and returns (data, headers, error).
func parseCSVString(csvData string) ([]map[string]interface{}, []string, error) {
	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(records) < 2 {
		return nil, nil, fmt.Errorf("CSV must have at least a header row and one data row")
	}
	headers := records[0]
	for i, h := range headers {
		headers[i] = strings.TrimSpace(h)
	}
	data := make([]map[string]interface{}, 0, len(records)-1)
	for _, record := range records[1:] {
		row := make(map[string]interface{}, len(headers))
		for i, h := range headers {
			if i < len(record) {
				v := strings.TrimSpace(record[i])
				if v != "" {
					row[h] = v
				}
			}
		}
		data = append(data, row)
	}
	return data, headers, nil
}

// --- Schema tools ---

func registerSchemaTools(s *mcp.Server, mctx *MCPContext) {
	type GetCollectionSchemaInput struct {
		CollectionID string `json:"collection_id" jsonschema:"ID of the collection"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_collection_schema",
		Description: "Get the property schema for a collection, which defines the typed fields for its objects.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input GetCollectionSchemaInput) (*mcp.CallToolResult, any, error) {
		user, token, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		resp, err := mctx.getCollectionsUC().Execute(ctx, usecases.GetCollectionsRequest{
			UserID:       user.ID(),
			CollectionID: &collectionID,
			UserToken:    token,
		})
		if err != nil || len(resp.Collections) == 0 {
			r, _ := errorResult(fmt.Errorf("collection not found"))
			return r, nil, nil
		}
		collection := resp.Collections[0]
		r, err := jsonResult(response.NewPropertySchemaResponse(collection.PropertySchema()))
		return r, nil, err
	})

	type PropertyDefinitionInput struct {
		Key          string `json:"key" jsonschema:"Snake_case storage key for the property"`
		DisplayName  string `json:"display_name" jsonschema:"Human-readable display name"`
		Type         string `json:"type" jsonschema:"Property type: text, currency, date, bool, url, numeric, grouped_text"`
		Required     bool   `json:"required,omitempty" jsonschema:"Whether this property is required"`
		CurrencyCode string `json:"currency_code,omitempty" jsonschema:"Currency code e.g. USD (only for currency type)"`
	}
	type UpdateCollectionSchemaInput struct {
		CollectionID string                    `json:"collection_id" jsonschema:"ID of the collection to update"`
		Definitions  []PropertyDefinitionInput `json:"definitions" jsonschema:"Property definitions for the schema"`
	}
	mcp.AddTool(s, &mcp.Tool{
		Name:        "update_collection_schema",
		Description: "Set or replace the property schema on a collection. This defines typed fields for object properties.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input UpdateCollectionSchemaInput) (*mcp.CallToolResult, any, error) {
		user, _, err := MCPUserFromContext(ctx)
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}

		collectionID, err := entities.CollectionIDFromString(input.CollectionID)
		if err != nil {
			r, _ := errorResult(fmt.Errorf("invalid collection_id: %w", err))
			return r, nil, nil
		}

		defs := make([]entities.PropertyDefinition, len(input.Definitions))
		for i, d := range input.Definitions {
			defs[i] = entities.PropertyDefinition{
				Key:          d.Key,
				DisplayName:  d.DisplayName,
				Type:         entities.PropertyType(d.Type),
				Required:     d.Required,
				CurrencyCode: d.CurrencyCode,
			}
		}
		schema := &entities.PropertySchema{Definitions: defs}

		resp, err := mctx.updatePropertySchemaUC().Execute(ctx, usecases.UpdatePropertySchemaRequest{
			CollectionID:   collectionID,
			UserID:         user.ID(),
			PropertySchema: schema,
		})
		if err != nil {
			r, _ := errorResult(err)
			return r, nil, nil
		}
		r, err := jsonResult(response.NewCollectionResponse(resp.Collection))
		return r, nil, err
	})
}
