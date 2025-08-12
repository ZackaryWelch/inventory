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

type CollectionController struct {
	createCollectionUC *usecases.CreateCollectionUseCase
	getCollectionsUC   *usecases.GetCollectionsUseCase
	updateCollectionUC *usecases.UpdateCollectionUseCase
	deleteCollectionUC *usecases.DeleteCollectionUseCase
	logger             *slog.Logger
}

func NewCollectionController(
	c *container.Container,
	logger *slog.Logger,
) *CollectionController {
	return &CollectionController{
		createCollectionUC: usecases.NewCreateCollectionUseCase(c.CollectionRepo, c.AuthService),
		getCollectionsUC:   usecases.NewGetCollectionsUseCase(c.CollectionRepo, c.AuthService),
		updateCollectionUC: usecases.NewUpdateCollectionUseCase(c.CollectionRepo, c.AuthService),
		deleteCollectionUC: usecases.NewDeleteCollectionUseCase(c.CollectionRepo),
		logger:             logger,
	}
}

// CreateCollection godoc
// @Summary Create a new collection
// @Description Create a new inventory collection
// @Tags collections
// @Accept json
// @Produce json
// @Param collection body request.CreateCollectionRequest true "Collection data"
// @Success 201 {object} response.CollectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/collections [post]
// @Security BearerAuth
func (ctrl *CollectionController) CreateCollection(c *gin.Context) {
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

	var req request.CreateCollectionRequest
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

	userID, err := req.GetUserID()
	if err != nil {
		ctrl.logger.Warn("Invalid user ID", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	// Validate user matches path parameter
	pathUserID, err := request.GetUserIDFromPath(c)
	if err != nil || !pathUserID.Equals(userID) || !pathUserID.Equals(user.ID()) {
		ctrl.logger.Warn("User ID mismatch", slog.String("path_user", pathUserID.String()), slog.String("req_user", userID.String()))
		c.JSON(http.StatusForbidden, gin.H{"error": "user ID mismatch"})
		return
	}

	groupID, err := req.GetGroupID()
	if err != nil {
		ctrl.logger.Warn("Invalid group ID", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
		return
	}

	ucReq := usecases.CreateCollectionRequest{
		UserID:     userID,
		GroupID:    groupID,
		Name:       req.Name,
		ObjectType: req.GetObjectType(),
		Tags:       req.Tags,
		Location:   req.Location,
		UserToken:  userToken,
	}

	resp, err := ctrl.createCollectionUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to create collection", slog.Any("error", err))
		if strings.Contains(err.Error(), "user is not a member of the group") {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create collection"})
		return
	}

	ctrl.logger.Info("Collection created successfully",
		slog.String("collection_id", resp.Collection.ID().String()),
		slog.String("collection_name", resp.Collection.Name().String()),
		slog.String("user_id", user.ID().String()))

	c.JSON(http.StatusCreated, response.NewCollectionResponse(resp.Collection))
}

// GetCollections godoc
// @Summary Get user's collections
// @Description Get all collections owned by or accessible to the user
// @Tags collections
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.CollectionListResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/collections [get]
// @Security BearerAuth
func (ctrl *CollectionController) GetCollections(c *gin.Context) {
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

	// Users can only access their own collections
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	ucReq := usecases.GetCollectionsRequest{
		UserID:    pathUserID,
		UserToken: userToken,
	}

	resp, err := ctrl.getCollectionsUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to get collections", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get collections"})
		return
	}

	ctrl.logger.Debug("Collections retrieved successfully",
		slog.String("user_id", user.ID().String()),
		slog.Int("collection_count", len(resp.Collections)))

	c.JSON(http.StatusOK, response.NewCollectionListResponse(resp.Collections))
}

// GetCollection godoc
// @Summary Get collection by ID
// @Description Get a specific collection by ID
// @Tags collections
// @Produce json
// @Param id path string true "User ID"
// @Param collection_id path string true "Collection ID"
// @Success 200 {object} response.CollectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/collections/{collection_id} [get]
// @Security BearerAuth
func (ctrl *CollectionController) GetCollection(c *gin.Context) {
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

	// Users can only access their own collections
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	ucReq := usecases.GetCollectionsRequest{
		UserID:       pathUserID,
		CollectionID: &collectionID,
		UserToken:    userToken,
	}

	resp, err := ctrl.getCollectionsUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to get collection", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get collection"})
		return
	}

	if len(resp.Collections) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
		return
	}

	ctrl.logger.Debug("Collection retrieved successfully",
		slog.String("collection_id", collectionID.String()),
		slog.String("user_id", user.ID().String()))

	c.JSON(http.StatusOK, response.NewCollectionResponse(resp.Collections[0]))
}

// UpdateCollection godoc
// @Summary Update collection
// @Description Update collection properties
// @Tags collections
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param collection_id path string true "Collection ID"
// @Param collection body request.UpdateCollectionRequest true "Collection update data"
// @Success 200 {object} response.CollectionResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/collections/{collection_id} [put]
// @Security BearerAuth
func (ctrl *CollectionController) UpdateCollection(c *gin.Context) {
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

	// Users can only update their own collections
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req request.UpdateCollectionRequest
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

	ucReq := usecases.UpdateCollectionRequest{
		CollectionID: collectionID,
		UserID:       pathUserID,
		Name:         &req.Name,
		Tags:         req.Tags,
		Location:     &req.Location,
		UserToken:    userToken,
	}

	resp, err := ctrl.updateCollectionUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to update collection", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update collection"})
		return
	}

	ctrl.logger.Info("Collection updated successfully",
		slog.String("collection_id", collectionID.String()),
		slog.String("user_id", user.ID().String()))

	c.JSON(http.StatusOK, response.NewCollectionResponse(resp.Collection))
}

// DeleteCollection godoc
// @Summary Delete collection
// @Description Delete a collection (only if empty)
// @Tags collections
// @Produce json
// @Param id path string true "User ID"
// @Param collection_id path string true "Collection ID"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /accounts/{id}/collections/{collection_id} [delete]
// @Security BearerAuth
func (ctrl *CollectionController) DeleteCollection(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
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

	// Users can only delete their own collections
	if !pathUserID.Equals(user.ID()) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	ucReq := usecases.DeleteCollectionRequest{
		CollectionID: collectionID,
		UserID:       pathUserID,
	}

	resp, err := ctrl.deleteCollectionUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to delete collection", slog.Any("error", err))
		if strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
			return
		}
		if strings.Contains(err.Error(), "cannot delete collection with objects") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete collection with objects"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete collection"})
		return
	}

	ctrl.logger.Info("Collection deleted successfully",
		slog.String("collection_id", collectionID.String()),
		slog.String("user_id", user.ID().String()))

	c.JSON(http.StatusOK, gin.H{"success": resp.Success})
}