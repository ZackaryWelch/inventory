package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidContainerID   = errors.New("invalid container ID")
	ErrInvalidContainerName = errors.New("container name must be between 1 and 255 characters")
	ErrInvalidContainerType = errors.New("invalid container type")
)

// ContainerType represents the type of physical container
type ContainerType string

const (
	ContainerTypeRoom      ContainerType = "room"
	ContainerTypeBookshelf ContainerType = "bookshelf"
	ContainerTypeShelf     ContainerType = "shelf"
	ContainerTypeBinder    ContainerType = "binder"
	ContainerTypeCabinet   ContainerType = "cabinet"
	ContainerTypeGeneral   ContainerType = "general" // Default for unspecified type
)

// ValidContainerTypes returns all valid container types
func ValidContainerTypes() []ContainerType {
	return []ContainerType{
		ContainerTypeRoom,
		ContainerTypeBookshelf,
		ContainerTypeShelf,
		ContainerTypeBinder,
		ContainerTypeCabinet,
		ContainerTypeGeneral,
	}
}

// IsValidContainerType checks if a string is a valid container type
func IsValidContainerType(t string) bool {
	switch ContainerType(t) {
	case ContainerTypeRoom, ContainerTypeBookshelf, ContainerTypeShelf,
		ContainerTypeBinder, ContainerTypeCabinet, ContainerTypeGeneral:
		return true
	}
	return false
}

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
	id                ContainerID
	collectionID      CollectionID
	name              ContainerName
	containerType     ContainerType // Type of container (room, bookshelf, shelf, etc.)
	parentContainerID *ContainerID  // Optional parent container for hierarchy
	categoryID        *CategoryID   // Optional category for this container
	groupID           *GroupID      // Optional group assignment for shared access
	objects           []Object      // Objects stored in this container
	location          string        // Physical location within collection
	// Physical dimensions for capacity planning
	width     *float64 // Width in inches
	depth     *float64 // Depth in inches
	rows      *int     // Number of rows/shelves
	capacity  *float64 // Total capacity in units
	createdAt time.Time
	updatedAt time.Time
}

type ContainerProps struct {
	CollectionID      CollectionID
	Name              ContainerName
	ContainerType     ContainerType
	ParentContainerID *ContainerID
	CategoryID        *CategoryID
	GroupID           *GroupID
	Location          string
	Width             *float64
	Depth             *float64
	Rows              *int
	Capacity          *float64
}

func NewContainer(props ContainerProps) (*Container, error) {
	// Default to general type if not specified
	containerType := props.ContainerType
	if containerType == "" {
		containerType = ContainerTypeGeneral
	}

	// Validate container type
	if !IsValidContainerType(string(containerType)) {
		return nil, ErrInvalidContainerType
	}

	now := time.Now()
	return &Container{
		id:                NewContainerID(),
		collectionID:      props.CollectionID,
		name:              props.Name,
		containerType:     containerType,
		parentContainerID: props.ParentContainerID,
		categoryID:        props.CategoryID,
		groupID:           props.GroupID,
		objects:           make([]Object, 0),
		location:          props.Location,
		width:             props.Width,
		depth:             props.Depth,
		rows:              props.Rows,
		capacity:          props.Capacity,
		createdAt:         now,
		updatedAt:         now,
	}, nil
}

func ReconstructContainer(id ContainerID, collectionID CollectionID, name ContainerName, containerType ContainerType, parentContainerID *ContainerID, categoryID *CategoryID, groupID *GroupID, objects []Object, location string, width, depth *float64, rows *int, capacity *float64, createdAt, updatedAt time.Time) *Container {
	// Default to general type if not specified
	if containerType == "" {
		containerType = ContainerTypeGeneral
	}

	return &Container{
		id:                id,
		collectionID:      collectionID,
		name:              name,
		containerType:     containerType,
		parentContainerID: parentContainerID,
		categoryID:        categoryID,
		groupID:           groupID,
		objects:           objects,
		location:          location,
		width:             width,
		depth:             depth,
		rows:              rows,
		capacity:          capacity,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
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

func (c *Container) GroupID() *GroupID {
	return c.groupID
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

func (c *Container) ContainerType() ContainerType {
	return c.containerType
}

func (c *Container) ParentContainerID() *ContainerID {
	return c.parentContainerID
}

func (c *Container) Width() *float64 {
	return c.width
}

func (c *Container) Depth() *float64 {
	return c.depth
}

func (c *Container) Rows() *int {
	return c.rows
}

func (c *Container) Capacity() *float64 {
	return c.capacity
}

// IsLeafContainer returns true if this container type cannot have children
func (c *Container) IsLeafContainer() bool {
	return c.containerType == ContainerTypeShelf ||
		c.containerType == ContainerTypeBinder ||
		c.containerType == ContainerTypeCabinet
}

// CanHaveChildren returns true if this container type can have child containers
func (c *Container) CanHaveChildren() bool {
	return c.containerType == ContainerTypeRoom ||
		c.containerType == ContainerTypeBookshelf ||
		c.containerType == ContainerTypeGeneral
}

// CalculateUsedCapacity calculates the currently used capacity based on objects
func (c *Container) CalculateUsedCapacity() float64 {
	// For now, each object counts as 1 unit
	// Can be enhanced to calculate based on object dimensions
	return float64(len(c.objects))
}

// GetCapacityUtilization returns the percentage of capacity used (0-100)
func (c *Container) GetCapacityUtilization() *float64 {
	if c.capacity == nil || *c.capacity == 0 {
		return nil
	}
	used := c.CalculateUsedCapacity()
	utilization := (used / *c.capacity) * 100
	return &utilization
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

func (c *Container) UpdateGroup(groupID *GroupID) error {
	c.groupID = groupID
	c.updatedAt = time.Now()
	return nil
}

func (c *Container) UpdateContainerType(containerType ContainerType) error {
	if !IsValidContainerType(string(containerType)) {
		return ErrInvalidContainerType
	}
	c.containerType = containerType
	c.updatedAt = time.Now()
	return nil
}

func (c *Container) UpdateParentContainer(parentContainerID *ContainerID) error {
	c.parentContainerID = parentContainerID
	c.updatedAt = time.Now()
	return nil
}

func (c *Container) UpdateDimensions(width, depth *float64, rows *int, capacity *float64) error {
	c.width = width
	c.depth = depth
	c.rows = rows
	c.capacity = capacity
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
