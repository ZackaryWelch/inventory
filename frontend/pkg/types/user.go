package types

import (
	"github.com/nishiki/backend-go/app/http/request"
	"github.com/nishiki/backend-go/app/http/response"
)

// Re-export backend response types
type User = response.UserResponse
type AuthInfoResponse = response.AuthInfoResponse
type ClaimsInfo = response.ClaimsInfo
type Group = response.GroupResponse
type Collection = response.CollectionResponse
type Container = response.ContainerResponse
type Object = response.ObjectResponse
type Category = response.CategoryResponse

// Re-export backend request types
type CreateGroupRequest = request.CreateGroupRequest
type UpdateGroupRequest = request.UpdateGroupRequest
type CreateCollectionRequest = request.CreateCollectionRequest
type UpdateCollectionRequest = request.UpdateCollectionRequest
type CreateContainerRequest = request.CreateContainerRequest
type UpdateContainerRequest = request.UpdateContainerRequest
type CreateObjectRequest = request.CreateObjectRequest
type UpdateObjectRequest = request.UpdateObjectRequest
type CreateCategoryRequest = request.CreateCategoryRequest
type UpdateCategoryRequest = request.UpdateCategoryRequest
