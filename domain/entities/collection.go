package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCollectionID     = errors.New("invalid collection ID")
	ErrInvalidCollectionName   = errors.New("collection name must be between 1 and 255 characters")
	ErrContainerNotFound       = errors.New("container not found in collection")
)

type CollectionID struct {
	value string
}

func NewCollectionID() CollectionID {
	return CollectionID{value: uuid.New().String()}
}

func CollectionIDFromString(id string) (CollectionID, error) {
	if id == "" {
		return CollectionID{}, ErrInvalidCollectionID
	}
	if _, err := uuid.Parse(id); err != nil {
		return CollectionID{}, ErrInvalidCollectionID
	}
	return CollectionID{value: id}, nil
}

func (c CollectionID) String() string {
	return c.value
}

func (c CollectionID) Equals(other CollectionID) bool {
	return c.value == other.value
}

type CollectionName struct {
	value string
}

func NewCollectionName(name string) (CollectionName, error) {
	if len(name) < 1 || len(name) > 255 {
		return CollectionName{}, ErrInvalidCollectionName
	}
	return CollectionName{value: name}, nil
}

func (c CollectionName) String() string {
	return c.value
}

func (c CollectionName) Equals(other CollectionName) bool {
	return c.value == other.value
}

type Collection struct {
	id          CollectionID
	userID      UserID       // Owner of the collection
	groupID     *GroupID     // Optional group for sharing this collection
	name        CollectionName
	categoryID  *CategoryID  // Optional category for this collection
	objectType  ObjectType   // Type of objects this collection holds
	containers  []Container  // Containers within this collection
	tags        []string
	location    string
	createdAt   time.Time
	updatedAt   time.Time
}

type CollectionProps struct {
	UserID     UserID
	GroupID    *GroupID
	Name       CollectionName
	CategoryID *CategoryID
	ObjectType ObjectType
	Tags       []string
	Location   string
}

func NewCollection(props CollectionProps) (*Collection, error) {
	now := time.Now()
	return &Collection{
		id:         NewCollectionID(),
		userID:     props.UserID,
		groupID:    props.GroupID,
		name:       props.Name,
		categoryID: props.CategoryID,
		objectType: props.ObjectType,
		containers: make([]Container, 0),
		tags:       props.Tags,
		location:   props.Location,
		createdAt:  now,
		updatedAt:  now,
	}, nil
}

func ReconstructCollection(id CollectionID, userID UserID, groupID *GroupID, name CollectionName, 
	categoryID *CategoryID, objectType ObjectType, containers []Container, tags []string, location string, 
	createdAt, updatedAt time.Time) *Collection {
	return &Collection{
		id:         id,
		userID:     userID,
		groupID:    groupID,
		name:       name,
		categoryID: categoryID,
		objectType: objectType,
		containers: containers,
		tags:       tags,
		location:   location,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

func (c *Collection) ID() CollectionID {
	return c.id
}

func (c *Collection) UserID() UserID {
	return c.userID
}

func (c *Collection) GroupID() *GroupID {
	return c.groupID
}

func (c *Collection) Name() CollectionName {
	return c.name
}

func (c *Collection) CategoryID() *CategoryID {
	return c.categoryID
}

func (c *Collection) ObjectType() ObjectType {
	return c.objectType
}

func (c *Collection) Containers() []Container {
	return append([]Container(nil), c.containers...)
}

func (c *Collection) Tags() []string {
	return append([]string(nil), c.tags...)
}

func (c *Collection) Location() string {
	return c.location
}

func (c *Collection) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Collection) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Collection) UpdateName(name CollectionName) error {
	c.name = name
	c.updatedAt = time.Now()
	return nil
}

func (c *Collection) AddContainer(container Container) error {
	c.containers = append(c.containers, container)
	c.updatedAt = time.Now()
	return nil
}

func (c *Collection) UpdateContainer(containerID ContainerID, updatedContainer Container) error {
	index := -1
	for i, container := range c.containers {
		if container.ID().Equals(containerID) {
			index = i
			break
		}
	}

	if index == -1 {
		return ErrContainerNotFound
	}

	c.containers[index] = updatedContainer
	c.updatedAt = time.Now()
	return nil
}

func (c *Collection) RemoveContainer(containerID ContainerID) error {
	index := -1
	for i, container := range c.containers {
		if container.ID().Equals(containerID) {
			index = i
			break
		}
	}

	if index == -1 {
		return ErrContainerNotFound
	}

	c.containers = append(c.containers[:index], c.containers[index+1:]...)
	c.updatedAt = time.Now()
	return nil
}

func (c *Collection) GetContainer(containerID ContainerID) (*Container, error) {
	for _, container := range c.containers {
		if container.ID().Equals(containerID) {
			return &container, nil
		}
	}
	return nil, ErrContainerNotFound
}

func (c *Collection) ContainerCount() int {
	return len(c.containers)
}

func (c *Collection) UpdateCategory(categoryID *CategoryID) error {
	c.categoryID = categoryID
	c.updatedAt = time.Now()
	return nil
}

func (c *Collection) HasCategory() bool {
	return c.categoryID != nil && !c.categoryID.IsZero()
}

func (c *Collection) GetAllObjects() []Object {
	var allObjects []Object
	for _, container := range c.containers {
		allObjects = append(allObjects, container.Objects()...)
	}
	return allObjects
}

func (c *Collection) GetObjectsByType(objectType ObjectType) []Object {
	var typeObjects []Object
	for _, container := range c.containers {
		typeObjects = append(typeObjects, container.GetObjectsByType(objectType)...)
	}
	return typeObjects
}

func (c *Collection) TotalObjectCount() int {
	count := 0
	for _, container := range c.containers {
		count += container.ObjectCount()
	}
	return count
}

func (c *Collection) Equals(other *Collection) bool {
	if other == nil {
		return false
	}
	return c.id.Equals(other.id)
}