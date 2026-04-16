package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/nishiki/backend/domain/entities"
	"github.com/nishiki/backend/domain/repositories"
	"github.com/nishiki/backend/domain/services"
)

type CreateObjectRequest struct {
	ContainerID   *entities.ContainerID  // nil = auto-assign to default container
	CollectionID  *entities.CollectionID // required when ContainerID is nil
	Name          string
	Description   string
	ObjectType    entities.ObjectType
	Location      string
	Quantity      *float64
	Unit          string
	Properties    map[string]entities.TypedValue // for direct callers (bulk import)
	RawProperties map[string]any                 // for HTTP/MCP callers; coerced in Execute()
	Tags          []string
	ExpiresAt     *time.Time
	UserID        entities.UserID
	UserToken     string
}

type CreateObjectResponse struct {
	Object      *entities.Object
	ContainerID entities.ContainerID
}

type CreateObjectUseCase struct {
	containerRepo  repositories.ContainerRepository
	collectionRepo repositories.CollectionRepository
	authService    services.AuthService
	typeInference  *services.TypeInferenceService
}

func NewCreateObjectUseCase(containerRepo repositories.ContainerRepository, collectionRepo repositories.CollectionRepository, authService services.AuthService) *CreateObjectUseCase {
	return &CreateObjectUseCase{
		containerRepo:  containerRepo,
		collectionRepo: collectionRepo,
		authService:    authService,
		typeInference:  services.NewTypeInferenceService(nil),
	}
}

func (uc *CreateObjectUseCase) Execute(ctx context.Context, req CreateObjectRequest) (*CreateObjectResponse, error) {
	var container *entities.Container
	var collection *entities.Collection
	var err error

	if req.ContainerID != nil {
		// Container specified — look it up
		container, err = uc.containerRepo.GetByID(ctx, *req.ContainerID)
		if err != nil {
			return nil, fmt.Errorf("container not found: %w", err)
		}
		collection, err = uc.collectionRepo.GetByID(ctx, container.CollectionID())
		if err != nil {
			return nil, fmt.Errorf("collection not found: %w", err)
		}
	} else if req.CollectionID != nil {
		// No container — find or create a default "General" container for the collection
		collection, err = uc.collectionRepo.GetByID(ctx, *req.CollectionID)
		if err != nil {
			return nil, fmt.Errorf("collection not found: %w", err)
		}
		container, err = uc.findOrCreateDefaultContainer(ctx, *req.CollectionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get default container: %w", err)
		}
	} else {
		return nil, errors.New("either container_id or collection_id is required")
	}

	// Check user access to collection
	userGroups, err := uc.authService.GetUserGroups(ctx, req.UserToken, req.UserID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	// Check access: user is owner OR user is member of collection's group
	hasAccess := collection.UserID().Equals(req.UserID)
	if !hasAccess && collection.GroupID() != nil {
		for _, group := range userGroups {
			if group.ID().Equals(*collection.GroupID()) {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil, errors.New("access denied: user does not have access to this collection")
	}

	// Create object name value object
	objectName, err := entities.NewObjectName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid object name: %w", err)
	}

	// Create object description
	objectDesc := entities.NewObjectDescription(req.Description)

	// Coerce raw properties from HTTP/MCP if provided
	props := req.Properties
	if len(req.RawProperties) > 0 {
		props = uc.typeInference.CoerceRawProperties(req.RawProperties, collection.PropertySchema())
	}

	// Create new object
	object, err := entities.NewObject(entities.ObjectProps{
		Name:        objectName,
		Description: objectDesc,
		ObjectType:  req.ObjectType,
		Location:    req.Location,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		Properties:  props,
		Tags:        req.Tags,
		ExpiresAt:   req.ExpiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create object entity: %w", err)
	}

	// Atomically add object to container using $push
	if err := uc.containerRepo.AddObject(ctx, container.ID(), *object); err != nil {
		return nil, fmt.Errorf("failed to add object to container: %w", err)
	}

	return &CreateObjectResponse{
		Object:      object,
		ContainerID: container.ID(),
	}, nil
}

const defaultContainerName = "General"

// findOrCreateDefaultContainer returns the default "General" container for a collection,
// creating one if it doesn't exist.
func (uc *CreateObjectUseCase) findOrCreateDefaultContainer(ctx context.Context, collectionID entities.CollectionID) (*entities.Container, error) {
	containers, err := uc.containerRepo.GetByCollectionID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	// Look for existing default container
	for _, c := range containers {
		if c.Name().String() == defaultContainerName && c.ContainerType() == entities.ContainerTypeGeneral {
			return c, nil
		}
	}

	// Create a new default container
	name, _ := entities.NewContainerName(defaultContainerName)
	container, err := entities.NewContainer(entities.ContainerProps{
		CollectionID:  collectionID,
		Name:          name,
		ContainerType: entities.ContainerTypeGeneral,
	})
	if err != nil {
		return nil, err
	}

	if err := uc.containerRepo.Create(ctx, container); err != nil {
		return nil, err
	}

	return container, nil
}
