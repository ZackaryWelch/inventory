package entities

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrInvalidObjectID   = errors.New("invalid object ID")
	ErrInvalidObjectName = errors.New("object name must be between 1 and 255 characters")
)

type ObjectID struct {
	value bson.ObjectID
}

func NewObjectID() ObjectID {
	return ObjectID{value: bson.NewObjectID()}
}

func ObjectIDFromHex(hex string) (ObjectID, error) {
	id, err := bson.ObjectIDFromHex(hex)
	if err != nil {
		return ObjectID{}, ErrInvalidObjectID
	}
	return ObjectID{value: id}, nil
}

func (o ObjectID) String() string {
	return o.value.Hex()
}

func (o ObjectID) ObjectID() bson.ObjectID {
	return o.value
}

func (o ObjectID) Equals(other ObjectID) bool {
	return o.value == other.value
}

func (o ObjectID) IsZero() bool {
	return o.value.IsZero()
}

type ObjectName struct {
	value string
}

func NewObjectName(name string) (ObjectName, error) {
	if len(name) < 1 || len(name) > 255 {
		return ObjectName{}, ErrInvalidObjectName
	}
	return ObjectName{value: name}, nil
}

func (o ObjectName) String() string {
	return o.value
}

func (o ObjectName) Equals(other ObjectName) bool {
	return o.value == other.value
}

type ObjectDescription struct {
	value string
}

func NewObjectDescription(description string) ObjectDescription {
	return ObjectDescription{value: description}
}

func (d ObjectDescription) String() string {
	return d.value
}

func (d ObjectDescription) Equals(other ObjectDescription) bool {
	return d.value == other.value
}

type ObjectType string

const (
	ObjectTypeFood      ObjectType = "food"
	ObjectTypeBook      ObjectType = "book"
	ObjectTypeVideoGame ObjectType = "videogame"
	ObjectTypeMusic     ObjectType = "music"
	ObjectTypeBoardGame ObjectType = "boardgame"
	ObjectTypeGeneral   ObjectType = "general"
)

func (ot ObjectType) String() string {
	return string(ot)
}

type Object struct {
	id          ObjectID
	name        ObjectName
	description ObjectDescription
	objectType  ObjectType
	quantity    *float64               // Optional quantity
	unit        string                 // Optional unit (e.g., "kg", "lbs", "pieces")
	properties  map[string]interface{} // Flexible properties for different object types
	tags        []string
	expiresAt   *time.Time // Optional expiration date (e.g., for food items)
	createdAt   time.Time
	updatedAt   time.Time
}

type ObjectProps struct {
	Name        ObjectName
	Description ObjectDescription
	ObjectType  ObjectType
	Quantity    *float64
	Unit        string
	Properties  map[string]interface{}
	Tags        []string
	ExpiresAt   *time.Time
}

func NewObject(props ObjectProps) (*Object, error) {
	now := time.Now()
	return &Object{
		id:          NewObjectID(),
		name:        props.Name,
		description: props.Description,
		objectType:  props.ObjectType,
		quantity:    props.Quantity,
		unit:        props.Unit,
		properties:  props.Properties,
		tags:        props.Tags,
		expiresAt:   props.ExpiresAt,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func ReconstructObject(id ObjectID, name ObjectName, description ObjectDescription, objectType ObjectType, quantity *float64, unit string, properties map[string]interface{}, tags []string, expiresAt *time.Time, createdAt, updatedAt time.Time) *Object {
	return &Object{
		id:          id,
		name:        name,
		description: description,
		objectType:  objectType,
		quantity:    quantity,
		unit:        unit,
		properties:  properties,
		tags:        tags,
		expiresAt:   expiresAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (o *Object) ID() ObjectID {
	return o.id
}

func (o *Object) Name() ObjectName {
	return o.name
}

func (o *Object) Description() ObjectDescription {
	return o.description
}

func (o *Object) ObjectType() ObjectType {
	return o.objectType
}

func (o *Object) Quantity() *float64 {
	return o.quantity
}

func (o *Object) Unit() string {
	return o.unit
}

func (o *Object) Properties() map[string]interface{} {
	if o.properties == nil {
		return make(map[string]interface{})
	}
	// Return a copy to prevent external modifications
	result := make(map[string]interface{})
	for k, v := range o.properties {
		result[k] = v
	}
	return result
}

func (o *Object) Tags() []string {
	return append([]string(nil), o.tags...)
}

func (o *Object) ExpiresAt() *time.Time {
	return o.expiresAt
}

func (o *Object) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Object) UpdatedAt() time.Time {
	return o.updatedAt
}

func (o *Object) GetProperty(key string) (interface{}, bool) {
	if o.properties == nil {
		return nil, false
	}
	value, exists := o.properties[key]
	return value, exists
}

func (o *Object) HasTag(tag string) bool {
	for _, t := range o.tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (o *Object) UpdateName(name ObjectName) error {
	o.name = name
	o.updatedAt = time.Now()
	return nil
}

func (o *Object) UpdateDescription(description ObjectDescription) error {
	o.description = description
	o.updatedAt = time.Now()
	return nil
}

func (o *Object) UpdateQuantity(quantity *float64) error {
	o.quantity = quantity
	o.updatedAt = time.Now()
	return nil
}

func (o *Object) UpdateUnit(unit string) error {
	o.unit = unit
	o.updatedAt = time.Now()
	return nil
}

func (o *Object) UpdateProperties(properties map[string]interface{}) error {
	o.properties = properties
	o.updatedAt = time.Now()
	return nil
}

func (o *Object) UpdateTags(tags []string) error {
	o.tags = tags
	o.updatedAt = time.Now()
	return nil
}

func (o *Object) UpdateExpiresAt(expiresAt *time.Time) error {
	o.expiresAt = expiresAt
	o.updatedAt = time.Now()
	return nil
}

func (o *Object) Equals(other *Object) bool {
	if other == nil {
		return false
	}
	return o.id.Equals(other.id)
}