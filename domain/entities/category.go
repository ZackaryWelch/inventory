package entities

import (
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrInvalidCategoryID   = errors.New("invalid category ID")
	ErrInvalidCategoryName = errors.New("category name must be between 1 and 50 characters")
)

type CategoryID struct {
	value bson.ObjectID
}

func NewCategoryID() CategoryID {
	return CategoryID{value: bson.NewObjectID()}
}

func CategoryIDFromObjectID(id bson.ObjectID) CategoryID {
	return CategoryID{value: id}
}

func CategoryIDFromHex(hex string) (CategoryID, error) {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		return CategoryID{}, ErrInvalidCategoryID
	}
	return CategoryID{value: id}, nil
}

func (id CategoryID) ObjectID() bson.ObjectID {
	return id.value
}

func (id CategoryID) Hex() string {
	return id.value.Hex()
}

func (id CategoryID) String() string {
	return id.value.Hex()
}

func (id CategoryID) IsZero() bool {
	return id.value.IsZero()
}

func (id CategoryID) Equals(other CategoryID) bool {
	return id.value == other.value
}

type CategoryName struct {
	value string
}

func NewCategoryName(value string) (CategoryName, error) {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) < 1 || len(trimmed) > 50 {
		return CategoryName{}, ErrInvalidCategoryName
	}
	return CategoryName{value: trimmed}, nil
}

func (n CategoryName) String() string {
	return n.value
}

func (n CategoryName) Equals(other CategoryName) bool {
	return n.value == other.value
}

type CategoryDescription struct {
	value string
}

func NewCategoryDescription(value string) CategoryDescription {
	return CategoryDescription{value: strings.TrimSpace(value)}
}

func (d CategoryDescription) String() string {
	return d.value
}

func (d CategoryDescription) IsEmpty() bool {
	return d.value == ""
}

func (d CategoryDescription) Equals(other CategoryDescription) bool {
	return d.value == other.value
}

type Category struct {
	id          CategoryID
	name        CategoryName
	description CategoryDescription
	createdAt   time.Time
	updatedAt   time.Time
}

type CategoryProps struct {
	Name        CategoryName
	Description CategoryDescription
}

func NewCategory(props CategoryProps) (*Category, error) {
	now := time.Now()
	return &Category{
		id:          NewCategoryID(),
		name:        props.Name,
		description: props.Description,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func ReconstructCategory(id CategoryID, name CategoryName, description CategoryDescription, createdAt, updatedAt time.Time) *Category {
	return &Category{
		id:          id,
		name:        name,
		description: description,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (c *Category) ID() CategoryID {
	return c.id
}

func (c *Category) Name() CategoryName {
	return c.name
}

func (c *Category) Description() CategoryDescription {
	return c.description
}

func (c *Category) CreatedAt() time.Time {
	return c.createdAt
}

func (c *Category) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *Category) UpdateName(name CategoryName) error {
	c.name = name
	c.updatedAt = time.Now()
	return nil
}

func (c *Category) UpdateDescription(description CategoryDescription) error {
	c.description = description
	c.updatedAt = time.Now()
	return nil
}

func (c *Category) Equals(other *Category) bool {
	if other == nil {
		return false
	}
	return c.id.Equals(other.id)
}
