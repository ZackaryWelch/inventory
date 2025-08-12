package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/nishiki/backend-go/domain/entities"
	"github.com/nishiki/backend-go/mocks"
)

func TestGetAllContainersUseCase_Execute(t *testing.T) {
	// Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockContainerRepo := mocks.NewMockContainerRepository(ctrl)
	mockAuthService := mocks.NewMockAuthService(ctrl)

	// Create use case
	useCase := NewGetAllContainersUseCase(mockContainerRepo, mockAuthService)

	// Test data
	ctx := context.Background()
	userID, _ := entities.UserIDFromString("test-user-123")
	userToken := "test-jwt-token"
	
	// Create test groups
	groupID1, _ := entities.GroupIDFromString("group-1")
	groupName1, _ := entities.NewGroupName("Test Group 1")
	group1 := entities.ReconstructGroup(groupID1, groupName1, []entities.ContainerID{}, []entities.UserID{}, time.Now(), time.Now())
	
	groupID2, _ := entities.GroupIDFromString("group-2") 
	groupName2, _ := entities.NewGroupName("Test Group 2")
	group2 := entities.ReconstructGroup(groupID2, groupName2, []entities.ContainerID{}, []entities.UserID{}, time.Now(), time.Now())

	// Create test containers
	containerID1, _ := entities.ContainerIDFromString("container-1")
	containerName1, _ := entities.NewContainerName("Test Container 1")
	container1 := entities.ReconstructContainer(containerID1, groupID1, containerName1, []entities.Food{}, time.Now(), time.Now())

	containerID2, _ := entities.ContainerIDFromString("container-2")  
	containerName2, _ := entities.NewContainerName("Test Container 2")
	container2 := entities.ReconstructContainer(containerID2, groupID1, containerName2, []entities.Food{}, time.Now(), time.Now())

	containerID3, _ := entities.ContainerIDFromString("container-3")
	containerName3, _ := entities.NewContainerName("Test Container 3") 
	container3 := entities.ReconstructContainer(containerID3, groupID2, containerName3, []entities.Food{}, time.Now(), time.Now())

	t.Run("Success - Returns containers from all user groups", func(t *testing.T) {
		// Mock auth service to return user groups
		mockAuthService.EXPECT().
			GetUserGroups(ctx, userToken, userID.String()).
			Return([]*entities.Group{group1, group2}, nil)

		// Mock container repo to return containers for each group
		mockContainerRepo.EXPECT().
			GetByGroupID(ctx, groupID1).
			Return([]*entities.Container{container1, container2}, nil)

		mockContainerRepo.EXPECT().
			GetByGroupID(ctx, groupID2).
			Return([]*entities.Container{container3}, nil)

		// Execute use case
		req := GetAllContainersRequest{
			UserID:    userID,
			UserToken: userToken,
		}

		resp, err := useCase.Execute(ctx, req)

		// Verify results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(resp.Containers) != 3 {
			t.Errorf("Expected 3 containers, got %d", len(resp.Containers))
		}

		// Check that all containers are present
		containerIDs := make(map[string]bool)
		for _, container := range resp.Containers {
			containerIDs[container.ID().String()] = true
		}

		expectedIDs := []string{containerID1.String(), containerID2.String(), containerID3.String()}
		for _, expectedID := range expectedIDs {
			if !containerIDs[expectedID] {
				t.Errorf("Expected container ID %s not found in results", expectedID)
			}
		}
	})

	t.Run("Success - No groups means no containers", func(t *testing.T) {
		// Mock auth service to return empty groups (user has no nishiki groups)
		mockAuthService.EXPECT().
			GetUserGroups(ctx, userToken, userID.String()).
			Return([]*entities.Group{}, nil)

		// Execute use case
		req := GetAllContainersRequest{
			UserID:    userID,
			UserToken: userToken,
		}

		resp, err := useCase.Execute(ctx, req)

		// Verify results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(resp.Containers) != 0 {
			t.Errorf("Expected 0 containers, got %d", len(resp.Containers))
		}
	})

	t.Run("Success - Group with no containers", func(t *testing.T) {
		// Mock auth service to return user groups
		mockAuthService.EXPECT().
			GetUserGroups(ctx, userToken, userID.String()).
			Return([]*entities.Group{group1}, nil)

		// Mock container repo to return empty containers for the group
		mockContainerRepo.EXPECT().
			GetByGroupID(ctx, groupID1).
			Return([]*entities.Container{}, nil)

		// Execute use case
		req := GetAllContainersRequest{
			UserID:    userID,
			UserToken: userToken,
		}

		resp, err := useCase.Execute(ctx, req)

		// Verify results
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if len(resp.Containers) != 0 {
			t.Errorf("Expected 0 containers, got %d", len(resp.Containers))
		}
	})

	t.Run("Error - Auth service fails to get user groups", func(t *testing.T) {
		// Mock auth service to return error
		mockAuthService.EXPECT().
			GetUserGroups(ctx, userToken, userID.String()).
			Return(nil, errors.New("failed to get groups"))

		// Execute use case
		req := GetAllContainersRequest{
			UserID:    userID,
			UserToken: userToken,
		}

		resp, err := useCase.Execute(ctx, req)

		// Verify error
		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		if resp != nil {
			t.Error("Expected nil response on error")
		}

		expectedError := "failed to get user groups"
		if !contains(err.Error(), expectedError) {
			t.Errorf("Expected error to contain '%s', got: %v", expectedError, err)
		}
	})

	t.Run("Partial Success - Container repo error for one group", func(t *testing.T) {
		// Mock auth service to return user groups
		mockAuthService.EXPECT().
			GetUserGroups(ctx, userToken, userID.String()).
			Return([]*entities.Group{group1, group2}, nil)

		// Mock container repo - first group succeeds, second group fails
		mockContainerRepo.EXPECT().
			GetByGroupID(ctx, groupID1).
			Return([]*entities.Container{container1, container2}, nil)

		mockContainerRepo.EXPECT().
			GetByGroupID(ctx, groupID2).
			Return(nil, errors.New("database connection failed"))

		// Execute use case
		req := GetAllContainersRequest{
			UserID:    userID,
			UserToken: userToken,
		}

		resp, err := useCase.Execute(ctx, req)

		// Verify results - should succeed with partial data
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		// Should only get containers from group1 (group2 failed)
		if len(resp.Containers) != 2 {
			t.Errorf("Expected 2 containers (from successful group), got %d", len(resp.Containers))
		}

		// Verify we got the right containers
		for _, container := range resp.Containers {
			if container.GroupID() != groupID1 {
				t.Errorf("Expected container from group1, got container from group %s", container.GroupID().String())
			}
		}
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}