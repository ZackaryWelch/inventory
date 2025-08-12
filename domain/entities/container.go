package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidContainerID    = errors.New("invalid container ID")
	ErrInvalidContainerName  = errors.New("container name must be between 1 and 255 characters")
)

type ContainerID struct {
	value string
}

func NewContainerID() ContainerID {
	return ContainerID{value: uuid.New().String()}
}

func ContainerIDFromString(id string) (ContainerID, error) {
	if id == "" {
		return ContainerID{}, ErrInvalidContainerID
	}
	if _, err := uuid.Parse(id); err != nil {
		return ContainerID{}, ErrInvalidContainerID
	}
	return ContainerID{value: id}, nil
}

func (c ContainerID) String() string {
	return c.value
}

func (c ContainerID) Equals(other ContainerID) bool {
	return c.value == other.value
}

type ContainerName struct {
	value string
}

func NewContainerName(name string) (ContainerName, error) {
	if len(name) < 1 || len(name) > 255 {
		return ContainerName{}, ErrInvalidContainerName
	}
	return ContainerName{value: name}, nil
}

func (c ContainerName) String() string {
	return c.value
}

func (c ContainerName) Equals(other ContainerName) bool {
	return c.value == other.value
}

type Container struct {
	id           ContainerID
	collectionID CollectionID
	name         ContainerName
	categoryID   *CategoryID // Optional category for this container
	objects      []Object    // Objects stored in this container
	location     string      // Physical location within collection
	createdAt    time.Time
	updatedAt    time.Time
}

type ContainerProps struct {
	CollectionID CollectionID
	Name         ContainerName
	CategoryID   *CategoryID
	Location     string
}

func NewContainer(props ContainerProps) (*Container, error) {
	now := time.Now()
	return &Container{
		id:           NewContainerID(),
		collectionID: props.CollectionID,
		name:         props.Name,
		categoryID:   props.CategoryID,
		objects:      make([]Object, 0),
		location:     props.Location,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func ReconstructContainer(id ContainerID, collectionID CollectionID, name ContainerName, categoryID *CategoryID, objects []Object, location string, createdAt, updatedAt time.Time) *Container {
	return &Container{
		id:           id,
		collectionID: collectionID,
		name:         name,
		categoryID:   categoryID,
		objects:      objects,
		location:     location,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (c *Container) ID() ContainerID {
	return c.id
}

func (c *Container) CollectionID() CollectionID {
	return c.collectionID
}

func (c *Container) Name() ContainerName {
	return c.name
}

func (c *Container) CategoryID() *CategoryID {
	return c.categoryID
}

func (c *Container) Objects() []Object {
	return append([]Object(nil), c.objects...)
}

func (c *Container) Location() string {
	return c.location
}

func (c *Container) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Container) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Container) UpdateName(name ContainerName) error {
	c.name = name
	c.updatedAt = time.Now()
	return nil
}

var (
	ErrObjectNotFoundInContainer = errors.New("object not found in container")
)

func (c *Container) AddObject(object Object) error {
	c.objects = append(c.objects, object)
	c.updatedAt = time.Now()
	return nil
}

func (c *Container) UpdateObject(objectID ObjectID, updatedObject Object) error {
	index := -1
	for i, object := range c.objects {
		if object.ID().Equals(objectID) {
			index = i
			break
		}
	}

	if index == -1 {
		return ErrObjectNotFoundInContainer
	}

	c.objects[index] = updatedObject
	c.updatedAt = time.Now()
	return nil
}

func (c *Container) RemoveObject(objectID ObjectID) error {
	index := -1
	for i, object := range c.objects {
		if object.ID().Equals(objectID) {
			index = i
			break
		}
	}

	if index == -1 {
		return ErrObjectNotFoundInContainer
	}

	// Remove object from slice
	c.objects = append(c.objects[:index], c.objects[index+1:]...)
	c.updatedAt = time.Now()
	return nil
}

func (c *Container) GetObject(objectID ObjectID) (*Object, error) {
	for _, object := range c.objects {
		if object.ID().Equals(objectID) {
			return &object, nil
		}
	}
	return nil, ErrObjectNotFoundInContainer
}

func (c *Container) ObjectCount() int {
	return len(c.objects)
}

func (c *Container) GetObjectsByType(objectType ObjectType) []Object {
	var typeObjects []Object
	for _, object := range c.objects {
		if object.ObjectType() == objectType {
			typeObjects = append(typeObjects, object)
		}
	}
	return typeObjects
}

func (c *Container) UpdateCategory(categoryID *CategoryID) error {
	c.categoryID = categoryID
	c.updatedAt = time.Now()
	return nil
}

func (c *Container) HasCategory() bool {
	return c.categoryID != nil && !c.categoryID.IsZero()
}

func (c *Container) Equals(other *Container) bool {
	if other == nil {
		return false
	}
	return c.id.Equals(other.id)
}
