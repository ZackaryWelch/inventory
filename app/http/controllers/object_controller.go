package controllers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/middleware"
	"github.com/nishiki/backend-go/app/http/request"
	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/usecases"
)

type ObjectController struct {
	createObjectUC           *usecases.CreateObjectUseCase
	updateObjectUC           *usecases.UpdateObjectUseCase
	deleteObjectUC           *usecases.DeleteObjectUseCase
	getCollectionObjectsUC   *usecases.GetCollectionObjectsUseCase
	bulkImportUC             *usecases.BulkImportObjectsUseCase
	logger                   *slog.Logger
}

func NewObjectController(
	c *container.Container,
	logger *slog.Logger,
) *ObjectController {
	return &ObjectController{
		createObjectUC:         usecases.NewCreateObjectUseCase(c.ContainerRepo, c.CollectionRepo, c.AuthService),
		updateObjectUC:         usecases.NewUpdateObjectUseCase(c.ContainerRepo, c.CollectionRepo, c.AuthService),
		deleteObjectUC:         usecases.NewDeleteObjectUseCase(c.ContainerRepo, c.CollectionRepo, c.AuthService),
		getCollectionObjectsUC: usecases.NewGetCollectionObjectsUseCase(c.CollectionRepo, c.ContainerRepo, c.AuthService),
		bulkImportUC:           usecases.NewBulkImportObjectsUseCase(c.ContainerRepo, c.CollectionRepo, c.AuthService),
		logger:                 logger,
	}
}

// CreateObject godoc
// @Summary Create a new object
// @Description Create a new object in a collection
// @Tags objects
// @Accept json
// @Produce json
// @Param object body request.CreateObjectRequest true "Object data"
// @Success 201 {object} response.ObjectResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/objects [post]
// @Security BearerAuth
func (ctrl *ObjectController) CreateObject(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(c)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	pathUserID, err := request.GetUserIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Users can only create objects in their own collections
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req request.CreateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.logger.Warn("Invalid request body", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		ctrl.logger.Warn("Request validation failed", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	containerID, err := req.GetContainerID()
	if err != nil {
		ctrl.logger.Warn("Invalid container ID", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container ID"})
		return
	}

	ucReq := usecases.CreateObjectRequest{
		ContainerID: containerID,
		Name:        req.Name,
		ObjectType:  req.GetObjectType(),
		Properties:  req.Properties,
		Tags:        req.Tags,
		UserID:      pathUserID,
		UserToken:   userToken,
	}

	resp, err := ctrl.createObjectUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to create object", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create object"})
		return
	}

	ctrl.logger.Info("Object created successfully",
		slog.String("object_id", resp.Object.ID().String()),
		slog.String("object_name", resp.Object.Name().String()),
		slog.String("container_id", containerID.String()),
		slog.String("user_id", user.ID().String()))

	c.JSON(http.StatusCreated, response.NewObjectResponse(*resp.Object))
}

// GetCollectionObjects godoc
// @Summary Get objects in collection
// @Description Get all objects in a specific collection
// @Tags objects
// @Produce json
// @Param id path string true "User ID"
// @Param collection_id path string true "Collection ID"
// @Success 200 {object} response.ObjectListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/collections/{collection_id}/objects [get]
// @Security BearerAuth
func (ctrl *ObjectController) GetCollectionObjects(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(c)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	pathUserID, err := request.GetUserIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collectionID, err := request.GetCollectionIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid collection ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Users can only access their own objects
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	ucReq := usecases.GetCollectionObjectsRequest{
		CollectionID: collectionID,
		UserID:       pathUserID,
		UserToken:    userToken,
	}

	resp, err := ctrl.getCollectionObjectsUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to get objects", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get objects"})
		return
	}

	ctrl.logger.Debug("Objects retrieved successfully",
		slog.String("collection_id", collectionID.String()),
		slog.String("user_id", user.ID().String()),
		slog.Int("object_count", len(resp.Objects)))

	c.JSON(http.StatusOK, response.NewObjectListResponse(resp.Objects))
}

// UpdateObject godoc
// @Summary Update object
// @Description Update object properties
// @Tags objects
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param object_id path string true "Object ID"
// @Param object body request.UpdateObjectRequest true "Object update data"
// @Success 200 {object} response.ObjectResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/objects/{object_id} [put]
// @Security BearerAuth
func (ctrl *ObjectController) UpdateObject(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(c)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	pathUserID, err := request.GetUserIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectID, err := request.GetObjectIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid object ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	containerID, err := request.GetContainerIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid container ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Users can only update their own objects
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req request.UpdateObjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.logger.Warn("Invalid request body", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		ctrl.logger.Warn("Request validation failed", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ucReq := usecases.UpdateObjectRequest{
		ContainerID: containerID,
		ObjectID:    objectID,
		Name:        req.Name,
		Properties:  req.Properties,
		Tags:        req.Tags,
		UserID:      pathUserID,
		UserToken:   userToken,
	}

	resp, err := ctrl.updateObjectUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to update object", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update object"})
		return
	}

	ctrl.logger.Info("Object updated successfully",
		slog.String("object_id", objectID.String()),
		slog.String("user_id", user.ID().String()))

	c.JSON(http.StatusOK, response.NewObjectResponse(*resp.Object))
}

// DeleteObject godoc
// @Summary Delete object
// @Description Delete an object from a collection
// @Tags objects
// @Produce json
// @Param id path string true "User ID"
// @Param object_id path string true "Object ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/objects/{object_id} [delete]
// @Security BearerAuth
func (ctrl *ObjectController) DeleteObject(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(c)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	pathUserID, err := request.GetUserIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectID, err := request.GetObjectIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid object ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	containerID, err := request.GetContainerIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid container ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Users can only delete their own objects
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	ucReq := usecases.DeleteObjectRequest{
		ContainerID: containerID,
		ObjectID:    objectID,
		UserID:      pathUserID,
		UserToken:   userToken,
	}

	resp, err := ctrl.deleteObjectUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to delete object", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "object not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete object"})
		return
	}

	ctrl.logger.Info("Object deleted successfully",
		slog.String("object_id", objectID.String()),
		slog.String("user_id", user.ID().String()))

	c.JSON(http.StatusOK, response.NewDeleteObjectResponse(resp.Success))
}

// BulkImport godoc
// @Summary Bulk import objects
// @Description Import multiple objects from JSON/CSV data
// @Tags objects
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param import body request.BulkImportRequest true "Bulk import data"
// @Success 200 {object} request.BulkImportResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/import [post]
// @Security BearerAuth
func (ctrl *ObjectController) BulkImport(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(c)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	pathUserID, err := request.GetUserIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	containerID, err := request.GetContainerIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid container ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Users can only import into their own containers
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req request.BulkImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.logger.Warn("Invalid request body", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		ctrl.logger.Warn("Request validation failed", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert raw data to ObjectImportData
	objectType := req.GetObjectType()
	objects := make([]usecases.ObjectImportData, len(req.Data))
	for i, item := range req.Data {
		// Extract name from the data
		name, ok := item["name"].(string)
		if !ok || name == "" {
			continue // Skip items without valid names
		}
		
		// Create properties from the remaining data
		properties := make(map[string]interface{})
		for key, value := range item {
			if key != "name" {
				properties[key] = value
			}
		}
		
		// Combine default tags with any item-specific tags
		tags := append([]string(nil), req.DefaultTags...)
		if itemTags, ok := item["tags"].([]string); ok {
			tags = append(tags, itemTags...)
		}
		
		objects[i] = usecases.ObjectImportData{
			Name:       name,
			ObjectType: objectType,
			Properties: properties,
			Tags:       tags,
		}
	}

	ucReq := usecases.BulkImportObjectsRequest{
		ContainerID: containerID,
		Objects:     objects,
		UserID:      pathUserID,
		UserToken:   userToken,
	}

	resp, err := ctrl.bulkImportUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to bulk import", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to import objects"})
		return
	}

	ctrl.logger.Info("Bulk import completed",
		slog.String("user_id", user.ID().String()),
		slog.Int("imported", resp.Imported),
		slog.Int("failed", resp.Failed))

	c.JSON(http.StatusOK, response.NewBulkImportResponse(resp.Imported, resp.Failed, resp.Total, resp.Errors))
}

// BulkImportToCollection godoc
// @Summary Bulk import objects to collection
// @Description Import multiple objects to a specific collection from JSON/CSV data
// @Tags objects
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param collection_id path string true "Collection ID"
// @Param import body request.BulkImportCollectionRequest true "Bulk import data"
// @Success 200 {object} request.BulkImportResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/collections/{collection_id}/import [post]
// @Security BearerAuth
func (ctrl *ObjectController) BulkImportToCollection(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(c)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	pathUserID, err := request.GetUserIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collectionID, err := request.GetCollectionIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid collection ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Users can only import into their own collections
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req request.BulkImportCollectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ctrl.logger.Warn("Invalid request body", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := req.Validate(); err != nil {
		ctrl.logger.Warn("Request validation failed", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use collection's object type (will be validated in use case)
	ucReq := usecases.BulkImportRequest{
		UserID:       pathUserID,
		CollectionID: &collectionID,
		Data:         req.Data,
		DefaultTags:  req.DefaultTags,
		UserToken:    userToken,
	}

	resp, err := ctrl.bulkImportUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to bulk import to collection", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to import objects"})
		return
	}

	ctrl.logger.Info("Bulk import to collection completed",
		slog.String("user_id", user.ID().String()),
		slog.String("collection_id", collectionID.String()),
		slog.Int("imported", resp.Imported),
		slog.Int("failed", resp.Failed))

	c.JSON(http.StatusOK, request.BulkImportResponse{
		Imported:  resp.Imported,
		Failed:    resp.Failed,
		Errors:    resp.Errors,
		ObjectIDs: resp.ObjectIDs,
	})
}