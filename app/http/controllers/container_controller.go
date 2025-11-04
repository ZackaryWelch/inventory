package controllers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/middleware"
	"github.com/nishiki/backend-go/app/http/request"
	"github.com/nishiki/backend-go/app/http/response"
	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/domain/usecases"
)

type ContainerController struct {
	createContainerUC           *usecases.CreateContainerUseCase
	updateContainerUC           *usecases.UpdateContainerUseCase
	getAllContainersUC          *usecases.GetAllContainersUseCase
	getContainerByIDUC          *usecases.GetContainerByIDUseCase
	getContainersUC             *usecases.GetContainersUseCase
	getContainersByCollectionUC *usecases.GetContainersByCollectionUseCase
	logger                      *slog.Logger
}

func NewContainerController(
	c *container.Container,
	logger *slog.Logger,
) *ContainerController {
	return &ContainerController{
		createContainerUC:           usecases.NewCreateContainerUseCase(c.ContainerRepo, c.CollectionRepo, c.AuthService),
		updateContainerUC:           usecases.NewUpdateContainerUseCase(c.ContainerRepo, c.CollectionRepo, c.AuthService),
		getAllContainersUC:          usecases.NewGetAllContainersUseCase(c.ContainerRepo, c.AuthService),
		getContainerByIDUC:          usecases.NewGetContainerByIDUseCase(c.ContainerRepo, c.CollectionRepo, c.AuthService),
		getContainersUC:             usecases.NewGetContainersUseCase(c.ContainerRepo, c.AuthService),
		getContainersByCollectionUC: usecases.NewGetContainersByCollectionUseCase(c.ContainerRepo, c.CollectionRepo, c.AuthService),
		logger:                      logger,
	}
}

// CreateContainer godoc
// @Summary Create a new container
// @Description Create a new food container in a group
// @Tags containers
// @Accept json
// @Produce json
// @Param container body request.CreateContainerRequest true "Container data"
// @Success 201 {object} response.ContainerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /containers [post]
// @Security BearerAuth
func (ctrl *ContainerController) CreateContainer(c *gin.Context) {
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

	var req request.CreateContainerRequest
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

	collectionID, err := req.GetCollectionID()
	if err != nil {
		ctrl.logger.Warn("Invalid collection ID", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection ID"})
		return
	}

	// Parse parent container ID if provided
	var parentContainerID *entities.ContainerID
	if req.ParentContainerID != nil {
		pid, err := entities.ContainerIDFromString(*req.ParentContainerID)
		if err != nil {
			ctrl.logger.Warn("Invalid parent container ID", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent container ID"})
			return
		}
		parentContainerID = &pid
	}

	// Parse group ID if provided
	var groupID *entities.GroupID
	if req.GroupID != nil {
		gid, err := entities.GroupIDFromString(*req.GroupID)
		if err != nil {
			ctrl.logger.Warn("Invalid group ID", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
			return
		}
		groupID = &gid
	}

	// Parse container type, default to general if not specified
	containerType := entities.ContainerTypeGeneral
	if req.Type != "" {
		containerType = entities.ContainerType(req.Type)
	}

	ucReq := usecases.CreateContainerRequest{
		CollectionID:      collectionID,
		Name:              req.Name,
		ContainerType:     containerType,
		ParentContainerID: parentContainerID,
		GroupID:           groupID,
		Location:          req.Location,
		Width:             req.Width,
		Depth:             req.Depth,
		Rows:              req.Rows,
		Capacity:          req.Capacity,
		UserID:            user.ID(),
		UserToken:         userToken,
	}

	resp, err := ctrl.createContainerUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to create container", slog.Any("error", err))
		if err.Error() == "user is not a member of the group" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create container"})
		return
	}

	ctrl.logger.Info("Container created successfully",
		slog.String("container_id", resp.Container.ID().String()),
		slog.String("container_name", resp.Container.Name().String()),
		slog.String("collection_id", collectionID.String()),
		slog.String("creator_id", user.ID().String()))

	c.JSON(http.StatusCreated, response.NewContainerResponse(resp.Container))
}

// GetContainers godoc
// @Summary Get all containers for user
// @Description Get all containers from groups the user is a member of
// @Tags containers
// @Produce json
// @Success 200 {object} response.ContainerListResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /containers [get]
// @Security BearerAuth
func (ctrl *ContainerController) GetContainers(c *gin.Context) {
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

	// Check if this is a nested route with collection_id
	collectionIDStr := c.Param("collection_id")
	if collectionIDStr != "" {
		// Get containers for specific collection
		collectionID, err := entities.CollectionIDFromString(collectionIDStr)
		if err != nil {
			ctrl.logger.Warn("Invalid collection ID", slog.Any("error", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection ID"})
			return
		}

		ucReq := usecases.GetContainersByCollectionRequest{
			CollectionID: collectionID,
			UserID:       user.ID(),
			UserToken:    userToken,
		}

		resp, err := ctrl.getContainersByCollectionUC.Execute(c.Request.Context(), ucReq)
		if err != nil {
			ctrl.logger.Error("Failed to get containers for collection", slog.Any("error", err))
			if err.Error() == "collection not found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "collection not found"})
				return
			}
			if err.Error() == "access denied: user does not have access to this collection" {
				c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get containers"})
			return
		}

		ctrl.logger.Debug("Containers retrieved successfully for collection",
			slog.String("collection_id", collectionID.String()),
			slog.String("user_id", user.ID().String()),
			slog.Int("container_count", len(resp.Containers)))

		c.JSON(http.StatusOK, response.NewContainerListResponse(resp.Containers))
		return
	}

	// Get all containers for user
	ucReq := usecases.GetAllContainersRequest{
		UserID:    user.ID(),
		UserToken: userToken,
	}

	resp, err := ctrl.getAllContainersUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to get containers", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get containers"})
		return
	}

	ctrl.logger.Debug("Containers retrieved successfully",
		slog.String("user_id", user.ID().String()),
		slog.Int("container_count", len(resp.Containers)))

	c.JSON(http.StatusOK, response.NewContainerListResponse(resp.Containers))
}

// GetContainer godoc
// @Summary Get container by ID
// @Description Get a specific container by ID
// @Tags containers
// @Produce json
// @Param id path string true "Container ID"
// @Success 200 {object} response.ContainerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /containers/{id} [get]
// @Security BearerAuth
func (ctrl *ContainerController) GetContainer(c *gin.Context) {
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

	containerID, err := request.GetContainerIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid container ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ucReq := usecases.GetContainerByIDRequest{
		ContainerID: containerID,
		UserID:      user.ID(),
		UserToken:   userToken,
	}

	resp, err := ctrl.getContainerByIDUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to get container", slog.Any("error", err))

		// Check if it's an access denied error
		if err.Error() == "access denied: user is not a member of the container's group" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}

		// Check if it's a not found error
		if err.Error() == "container not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get container"})
		return
	}

	ctrl.logger.Debug("Container retrieved successfully",
		slog.String("container_id", containerID.String()),
		slog.String("user_id", user.ID().String()))

	c.JSON(http.StatusOK, response.NewContainerResponse(resp.Container))
}

// UpdateContainer godoc
// @Summary Update a container
// @Description Update an existing container's properties
// @Tags containers
// @Accept json
// @Produce json
// @Param id path string true "Container ID"
// @Param container body request.UpdateContainerRequest true "Container data"
// @Success 200 {object} response.ContainerResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /containers/{id} [put]
// @Security BearerAuth
func (ctrl *ContainerController) UpdateContainer(c *gin.Context) {
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

	containerID, err := request.GetContainerIDFromPath(c)
	if err != nil {
		ctrl.logger.Warn("Invalid container ID in path", slog.Any("error", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req request.UpdateContainerRequest
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

	// Build use case request
	ucReq := usecases.UpdateContainerRequest{
		ContainerID: containerID,
		UserID:      user.ID(),
		UserToken:   userToken,
	}

	// Only set fields that are provided
	if req.Name != "" {
		ucReq.Name = &req.Name
	}

	if req.Type != "" {
		containerType := entities.ContainerType(req.Type)
		ucReq.ContainerType = &containerType
	}

	if req.ParentContainerID != nil {
		if *req.ParentContainerID == "" {
			// Empty string means remove parent
			var nilParent *entities.ContainerID
			ucReq.ParentContainerID = &nilParent
		} else {
			parentID, err := entities.ContainerIDFromString(*req.ParentContainerID)
			if err != nil {
				ctrl.logger.Warn("Invalid parent container ID", slog.Any("error", err))
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent container ID"})
				return
			}
			parentIDPtr := &parentID
			ucReq.ParentContainerID = &parentIDPtr
		}
	}

	if req.GroupID != nil {
		if *req.GroupID == "" {
			// Empty string means remove group
			var nilGroup *entities.GroupID
			ucReq.GroupID = &nilGroup
		} else {
			groupID, err := entities.GroupIDFromString(*req.GroupID)
			if err != nil {
				ctrl.logger.Warn("Invalid group ID", slog.Any("error", err))
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid group ID"})
				return
			}
			groupIDPtr := &groupID
			ucReq.GroupID = &groupIDPtr
		}
	}

	if req.Location != "" {
		ucReq.Location = &req.Location
	}

	if req.Width != nil {
		ucReq.Width = &req.Width
	}

	if req.Depth != nil {
		ucReq.Depth = &req.Depth
	}

	if req.Rows != nil {
		ucReq.Rows = &req.Rows
	}

	if req.Capacity != nil {
		ucReq.Capacity = &req.Capacity
	}

	resp, err := ctrl.updateContainerUC.Execute(c.Request.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to update container", slog.Any("error", err))
		if err.Error() == "container not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "container not found"})
			return
		}
		if err.Error() == "access denied: user does not have access to this container" {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update container"})
		return
	}

	ctrl.logger.Info("Container updated successfully",
		slog.String("container_id", containerID.String()),
		slog.String("user_id", user.ID().String()))

	c.JSON(http.StatusOK, response.NewContainerResponse(resp.Container))
}
