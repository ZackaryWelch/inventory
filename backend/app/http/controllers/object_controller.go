package controllers

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/httputil"
	"github.com/nishiki/backend-go/app/http/middleware"
	"github.com/nishiki/backend-go/app/http/request"
	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/usecases"
)

type ObjectController struct {
	createObjectUC         *usecases.CreateObjectUseCase
	updateObjectUC         *usecases.UpdateObjectUseCase
	deleteObjectUC         *usecases.DeleteObjectUseCase
	getCollectionObjectsUC *usecases.GetCollectionObjectsUseCase
	bulkImportUC           *usecases.BulkImportObjectsUseCase
	bulkImportCollectionUC *usecases.BulkImportCollectionUseCase
	logger                 *slog.Logger
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
		bulkImportCollectionUC: usecases.NewBulkImportCollectionUseCase(c.CollectionRepo, c.ContainerRepo, c.AuthService, c.GetConfig().Import.ReservedColumns),
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
func (ctrl *ObjectController) CreateObject(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.GetCurrentUser(r)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(r)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	pathUserID, err := request.GetUserIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Users can only create objects in their own collections
	if !pathUserID.Equals(user.ID()) {
		httputil.Error(w, http.StatusForbidden, "access denied")
		return
	}

	var req request.CreateObjectRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		ctrl.logger.Warn("Invalid request body", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		ctrl.logger.Warn("Request validation failed", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	containerID, err := req.GetContainerID()
	if err != nil {
		ctrl.logger.Warn("Invalid container ID", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, "invalid container ID")
		return
	}

	ucReq := usecases.CreateObjectRequest{
		ContainerID: containerID,
		Name:        req.Name,
		Description: req.Description,
		ObjectType:  req.GetObjectType(),
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		Properties:  req.Properties,
		Tags:        req.Tags,
		ExpiresAt:   req.ExpiresAt,
		UserID:      pathUserID,
		UserToken:   userToken,
	}

	resp, err := ctrl.createObjectUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to create object", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			httputil.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			httputil.Error(w, http.StatusNotFound, "container not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to create object")
		return
	}

	ctrl.logger.Info("Object created successfully",
		slog.String("object_id", resp.Object.ID().String()),
		slog.String("object_name", resp.Object.Name().String()),
		slog.String("container_id", containerID.String()),
		slog.String("user_id", user.ID().String()))

	httputil.JSON(w, http.StatusCreated, response.NewObjectResponse(*resp.Object, containerID.String()))
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
func (ctrl *ObjectController) GetCollectionObjects(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.GetCurrentUser(r)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(r)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	pathUserID, err := request.GetUserIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	collectionID, err := request.GetCollectionIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid collection ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Users can only access their own objects
	if !pathUserID.Equals(user.ID()) {
		httputil.Error(w, http.StatusForbidden, "access denied")
		return
	}

	ucReq := usecases.GetCollectionObjectsRequest{
		CollectionID: collectionID,
		UserID:       pathUserID,
		UserToken:    userToken,
	}

	// Parse optional filter query parameters.
	q := r.URL.Query()
	ucReq.Query = q.Get("q")
	ucReq.Tags = q["tag"]

	if cidStr := q.Get("container_id"); cidStr != "" {
		cid, err := entities.ContainerIDFromString(cidStr)
		if err != nil {
			ctrl.logger.Warn("Invalid container_id query param", slog.Any("error", err))
			httputil.Error(w, http.StatusBadRequest, "invalid container_id")
			return
		}
		ucReq.ContainerID = &cid
	}

	// Parse property[key]=value filters.
	for paramKey, values := range q {
		if strings.HasPrefix(paramKey, "property[") && strings.HasSuffix(paramKey, "]") {
			propKey := paramKey[len("property[") : len(paramKey)-1]
			if ucReq.PropertyFilters == nil {
				ucReq.PropertyFilters = make(map[string]string)
			}
			ucReq.PropertyFilters[propKey] = values[0]
		}
	}

	resp, err := ctrl.getCollectionObjectsUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to get objects", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			httputil.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			httputil.Error(w, http.StatusNotFound, "collection not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get objects")
		return
	}
	ctrl.logger.Debug("Objects retrieved successfully",
		slog.String("collection_id", collectionID.String()),
		slog.String("user_id", user.ID().String()),
		slog.Int("object_count", len(resp.Objects)))

	objectResponses := make([]response.ObjectResponse, len(resp.Objects))
	for i, item := range resp.Objects {
		objectResponses[i] = response.NewObjectResponse(item.Object, item.ContainerID.String())
	}
	httputil.JSON(w, http.StatusOK, response.ObjectListResponse{Objects: objectResponses, Total: len(objectResponses)})
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
func (ctrl *ObjectController) UpdateObject(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.GetCurrentUser(r)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(r)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	pathUserID, err := request.GetUserIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	objectID, err := request.GetObjectIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid object ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Users can only update their own objects
	if !pathUserID.Equals(user.ID()) {
		httputil.Error(w, http.StatusForbidden, "access denied")
		return
	}

	var req request.UpdateObjectRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		ctrl.logger.Warn("Invalid request body", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		ctrl.logger.Warn("Request validation failed", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	containerID, err := req.GetContainerID()
	if err != nil {
		ctrl.logger.Warn("Invalid container ID in body", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, "invalid container_id")
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

	resp, err := ctrl.updateObjectUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to update object", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			httputil.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			httputil.Error(w, http.StatusNotFound, "object not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to update object")
		return
	}

	ctrl.logger.Info("Object updated successfully",
		slog.String("object_id", objectID.String()),
		slog.String("user_id", user.ID().String()))

	httputil.JSON(w, http.StatusOK, response.NewObjectResponse(*resp.Object, containerID.String()))
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
func (ctrl *ObjectController) DeleteObject(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.GetCurrentUser(r)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(r)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	pathUserID, err := request.GetUserIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	objectID, err := request.GetObjectIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid object ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Users can only delete their own objects
	if !pathUserID.Equals(user.ID()) {
		httputil.Error(w, http.StatusForbidden, "access denied")
		return
	}

	// container_id is optional — use case looks it up automatically if omitted
	var containerID *entities.ContainerID
	if cidStr := r.URL.Query().Get("container_id"); cidStr != "" {
		cid, err := entities.ContainerIDFromString(cidStr)
		if err != nil {
			ctrl.logger.Warn("Invalid container ID", slog.Any("error", err))
			httputil.Error(w, http.StatusBadRequest, "invalid container_id")
			return
		}
		containerID = &cid
	}

	ucReq := usecases.DeleteObjectRequest{
		ContainerID: containerID,
		ObjectID:    objectID,
		UserID:      pathUserID,
		UserToken:   userToken,
	}

	resp, err := ctrl.deleteObjectUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to delete object", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			httputil.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			httputil.Error(w, http.StatusNotFound, "object not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to delete object")
		return
	}

	ctrl.logger.Info("Object deleted successfully",
		slog.String("object_id", objectID.String()),
		slog.String("user_id", user.ID().String()))

	httputil.JSON(w, http.StatusOK, response.DeleteObjectResponse{
		Success: resp.Success,
	})
}

// RemoveObjectFromContainer godoc
// @Summary Remove object from a specific container
// @Description Remove an object from a specific container (container ID required in path)
// @Tags objects
// @Produce json
// @Param id path string true "User ID"
// @Param collection_id path string true "Collection ID"
// @Param container_id path string true "Container ID"
// @Param object_id path string true "Object ID"
// @Success 200 {object} response.DeleteObjectResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/collections/{collection_id}/containers/{container_id}/objects/{object_id} [delete]
// @Security BearerAuth
func (ctrl *ObjectController) RemoveObjectFromContainer(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.GetCurrentUser(r)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(r)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	pathUserID, err := request.GetUserIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	containerID, err := request.GetContainerIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid container ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	objectID, err := request.GetObjectIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid object ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if !pathUserID.Equals(user.ID()) {
		httputil.Error(w, http.StatusForbidden, "access denied")
		return
	}

	ucReq := usecases.DeleteObjectRequest{
		ContainerID: &containerID,
		ObjectID:    objectID,
		UserID:      pathUserID,
		UserToken:   userToken,
	}

	resp, err := ctrl.deleteObjectUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to remove object from container", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			httputil.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			httputil.Error(w, http.StatusNotFound, "object or container not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to remove object")
		return
	}

	ctrl.logger.Info("Object removed from container",
		slog.String("object_id", objectID.String()),
		slog.String("container_id", containerID.String()),
		slog.String("user_id", user.ID().String()))

	httputil.JSON(w, http.StatusOK, response.DeleteObjectResponse{
		Success: resp.Success,
	})
}

// BulkImport godoc
// @Summary Bulk import objects to a container
// @Description Import multiple objects into a specific container. container_id must be provided in the request body.
// @Tags objects
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param import body request.BulkImportRequest true "Bulk import data (must include container_id)"
// @Success 200 {object} response.BulkImportResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/import [post]
// @Security BearerAuth
func (ctrl *ObjectController) BulkImport(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.GetCurrentUser(r)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(r)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	pathUserID, err := request.GetUserIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Users can only import into their own containers
	if !pathUserID.Equals(user.ID()) {
		httputil.Error(w, http.StatusForbidden, "access denied")
		return
	}

	var req request.BulkImportRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		ctrl.logger.Warn("Invalid request body", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		ctrl.logger.Warn("Request validation failed", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	containerID, err := req.GetContainerID()
	if err != nil {
		ctrl.logger.Warn("Invalid container ID in body", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, "invalid container_id")
		return
	}

	objectType := req.GetObjectType()
	objects := make([]usecases.ObjectImportData, len(req.Data))
	for i, item := range req.Data {
		name, ok := item["name"].(string)
		if !ok || name == "" {
			continue
		}

		properties := make(map[string]interface{})
		for key, value := range item {
			if key != "name" {
				properties[key] = value
			}
		}

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

	resp, err := ctrl.bulkImportUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to bulk import", slog.Any("error", err))
		httputil.Error(w, http.StatusInternalServerError, "failed to import objects")
		return
	}

	ctrl.logger.Info("Bulk import completed",
		slog.String("user_id", user.ID().String()),
		slog.String("container_id", containerID.String()),
		slog.Int("imported", resp.Imported),
		slog.Int("failed", resp.Failed))

	httputil.JSON(w, http.StatusOK, response.BulkImportResponse{
		Imported: resp.Imported,
		Failed:   resp.Failed,
		Total:    resp.Total,
		Errors:   resp.Errors,
	})
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
func (ctrl *ObjectController) BulkImportToCollection(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.GetCurrentUser(r)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	userToken, tokenExists := middleware.GetCurrentToken(r)
	if !tokenExists {
		ctrl.logger.Error("No auth token found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	pathUserID, err := request.GetUserIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid user ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	collectionID, err := request.GetCollectionIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid collection ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Users can only import into their own collections
	if !pathUserID.Equals(user.ID()) {
		httputil.Error(w, http.StatusForbidden, "access denied")
		return
	}

	var req request.BulkImportCollectionRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		ctrl.logger.Warn("Invalid request body", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		ctrl.logger.Warn("Request validation failed", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// Parse target container ID if provided
	var targetContainerID *entities.ContainerID
	if req.TargetContainerID != nil {
		cID, err := entities.ContainerIDFromString(*req.TargetContainerID)
		if err != nil {
			ctrl.logger.Warn("Invalid target container ID", slog.Any("error", err))
			httputil.Error(w, http.StatusBadRequest, "invalid target_container_id")
			return
		}
		targetContainerID = &cID
	}

	// Use collection's object type (will be validated in use case)
	ucReq := usecases.BulkImportCollectionRequest{
		UserID:            pathUserID,
		CollectionID:      collectionID,
		TargetContainerID: targetContainerID,
		DistributionMode:  req.DistributionMode,
		Data:              req.Data,
		DefaultTags:       req.DefaultTags,
		UserToken:         userToken,
		LocationColumn:    req.LocationColumn,
		NameColumn:        req.NameColumn,
		InferSchema:       req.InferSchema,
	}

	resp, err := ctrl.bulkImportCollectionUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to bulk import to collection", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			httputil.Error(w, http.StatusForbidden, "access denied")
			return
		}
		if strings.Contains(err.Error(), "not found") {
			httputil.Error(w, http.StatusNotFound, "collection not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to import objects")
		return
	}

	ctrl.logger.Info("Bulk import to collection completed",
		slog.String("user_id", user.ID().String()),
		slog.String("collection_id", collectionID.String()),
		slog.Int("imported", resp.Imported),
		slog.Int("failed", resp.Failed))

	httputil.JSON(w, http.StatusOK, response.BulkImportResponse{
		Imported: resp.Imported,
		Failed:   resp.Failed,
		Total:    resp.Total,
		Errors:   resp.Errors,
	})
}
