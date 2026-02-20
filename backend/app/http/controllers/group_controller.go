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
	"github.com/nishiki/backend-go/domain/services"
	"github.com/nishiki/backend-go/domain/usecases"
)

type GroupController struct {
	createGroupUC   *usecases.CreateGroupUseCase
	getGroupsUC     *usecases.GetGroupsUseCase
	getContainersUC *usecases.GetContainersUseCase
	authService     services.AuthService
	logger          *slog.Logger
}

func NewGroupController(
	c *container.Container,
	logger *slog.Logger,
) *GroupController {
	return &GroupController{
		createGroupUC:   usecases.NewCreateGroupUseCase(c.AuthService),
		getGroupsUC:     usecases.NewGetGroupsUseCase(c.AuthService),
		getContainersUC: usecases.NewGetContainersUseCase(c.ContainerRepo, c.AuthService),
		authService:     c.AuthService,
		logger:          logger,
	}
}

// CreateGroup godoc
// @Summary Create a new group
// @Description Create a new food storage group
// @Tags groups
// @Accept json
// @Produce json
// @Param group body request.CreateGroupRequest true "Group data"
// @Success 201 {object} response.GroupResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups [post]
// @Security BearerAuth
func (ctrl *GroupController) CreateGroup(w http.ResponseWriter, r *http.Request) {
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

	var req request.CreateGroupRequest
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

	ucReq := usecases.CreateGroupRequest{
		Name:      req.Name,
		CreatorID: user.ID(),
		UserToken: userToken,
	}

	resp, err := ctrl.createGroupUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to create group", slog.Any("error", err))
		// Check if it's an authentication failure
		if strings.Contains(err.Error(), "authentication failed") || strings.Contains(err.Error(), "invalid token") {
			httputil.Error(w, http.StatusUnauthorized, "authentication failed")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to create group")
		return
	}

	ctrl.logger.Info("Group created successfully",
		slog.String("group_id", resp.Group.ID().String()),
		slog.String("group_name", resp.Group.Name().String()),
		slog.String("creator_id", user.ID().String()))

	httputil.JSON(w, http.StatusCreated, response.NewGroupResponse(resp.Group))
}

// GetGroups godoc
// @Summary Get user's groups
// @Description Get all groups that the authenticated user is a member of
// @Tags groups
// @Produce json
// @Success 200 {object} response.GroupListResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups [get]
// @Security BearerAuth
func (ctrl *GroupController) GetGroups(w http.ResponseWriter, r *http.Request) {
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

	ucReq := usecases.GetGroupsRequest{
		UserID:    user.ID(),
		UserToken: userToken,
	}

	resp, err := ctrl.getGroupsUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to get groups", slog.Any("error", err))
		// Check if it's an authentication failure
		if strings.Contains(err.Error(), "authentication failed") || strings.Contains(err.Error(), "invalid token") {
			httputil.Error(w, http.StatusUnauthorized, "authentication failed")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get groups")
		return
	}

	ctrl.logger.Debug("Groups retrieved successfully",
		slog.String("user_id", user.ID().String()),
		slog.Int("group_count", len(resp.Groups)))

	httputil.JSON(w, http.StatusOK, response.NewGroupListResponse(resp.Groups))
}

// GetGroupContainers godoc
// @Summary Get containers in a group
// @Description Get all containers in a specific group
// @Tags groups
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} response.ContainerListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id}/containers [get]
// @Security BearerAuth
func (ctrl *GroupController) GetGroupContainers(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := request.GetGroupIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid group ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	ucReq := usecases.GetContainersRequest{
		GroupID:   groupID,
		UserID:    user.ID(),
		UserToken: userToken,
	}

	resp, err := ctrl.getContainersUC.Execute(r.Context(), ucReq)
	if err != nil {
		ctrl.logger.Error("Failed to get containers", slog.Any("error", err))
		if err.Error() == "user is not a member of the group" {
			httputil.Error(w, http.StatusForbidden, "access denied")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get containers")
		return
	}

	ctrl.logger.Debug("Containers retrieved successfully",
		slog.String("group_id", groupID.String()),
		slog.String("user_id", user.ID().String()),
		slog.Int("container_count", len(resp.Containers)))

	httputil.JSON(w, http.StatusOK, response.NewContainerListResponse(resp.Containers))
}

// GetGroup godoc
// @Summary Get group by ID
// @Description Get a specific group by ID
// @Tags groups
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} response.GroupResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id} [get]
// @Security BearerAuth
func (ctrl *GroupController) GetGroup(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := request.GetGroupIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid group ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	group, err := ctrl.authService.GetGroupByID(r.Context(), userToken, groupID.String())
	if err != nil {
		ctrl.logger.Error("Failed to get group", slog.Any("error", err))
		if err.Error() == "group not found" {
			httputil.Error(w, http.StatusNotFound, "group not found")
			return
		}
		httputil.Error(w, http.StatusInternalServerError, "failed to get group")
		return
	}

	ctrl.logger.Debug("Group retrieved successfully",
		slog.String("group_id", groupID.String()),
		slog.String("user_id", user.ID().String()))

	httputil.JSON(w, http.StatusOK, response.NewGroupResponse(group))
}

// GetGroupUsers godoc
// @Summary Get users in a group
// @Description Get all users that are members of a specific group
// @Tags groups
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} response.UserListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/{id}/users [get]
// @Security BearerAuth
func (ctrl *GroupController) GetGroupUsers(w http.ResponseWriter, r *http.Request) {
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

	groupID, err := request.GetGroupIDFromPath(r)
	if err != nil {
		ctrl.logger.Warn("Invalid group ID in path", slog.Any("error", err))
		httputil.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	// TODO: Check if user is a member of the group for authorization
	// For now, we'll allow any authenticated user to see group members

	users, err := ctrl.authService.GetGroupUsers(r.Context(), userToken, groupID.String())
	if err != nil {
		ctrl.logger.Error("Failed to get group users", slog.Any("error", err))
		httputil.Error(w, http.StatusInternalServerError, "failed to get group users")
		return
	}

	ctrl.logger.Debug("Group users retrieved successfully",
		slog.String("group_id", groupID.String()),
		slog.String("user_id", user.ID().String()),
		slog.Int("user_count", len(users)))

	httputil.JSON(w, http.StatusOK, response.NewUserListResponse(users))
}

// JoinGroup godoc
// @Summary Join a group
// @Description Join a group using an invitation hash
// @Tags groups
// @Accept json
// @Produce json
// @Param join body request.JoinGroupRequest true "Join group data"
// @Success 200 {object} response.JoinGroupResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /groups/join [post]
// @Security BearerAuth
func (ctrl *GroupController) JoinGroup(w http.ResponseWriter, r *http.Request) {
	user, exists := middleware.GetCurrentUser(r)
	if !exists {
		ctrl.logger.Error("No authenticated user found in context")
		httputil.Error(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req request.JoinGroupRequest
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

	// TODO: Implement invitation hash lookup and group joining logic
	// For now, this is a placeholder implementation
	ctrl.logger.Info("Group join attempted",
		slog.String("invitation_hash", req.InvitationHash),
		slog.String("user_id", user.ID().String()))

	// Placeholder response
	httputil.Error(w, http.StatusNotImplemented, "group join not implemented yet")
}
