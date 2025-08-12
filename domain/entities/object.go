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
	id         ObjectID
	name       ObjectName
	objectType ObjectType
	properties map[string]interface{} // Flexible properties for different object types
	tags       []string
	createdAt  time.Time
}

type ObjectProps struct {
	Name       ObjectName
	ObjectType ObjectType
	Properties map[string]interface{}
	Tags       []string
}

func NewObject(props ObjectProps) (*Object, error) {
	return &Object{
		id:         NewObjectID(),
		name:       props.Name,
		objectType: props.ObjectType,
		properties: props.Properties,
		tags:       props.Tags,
		createdAt:  time.Now(),
	}, nil
}

func ReconstructObject(id ObjectID, name ObjectName, objectType ObjectType, properties map[string]interface{}, tags []string, createdAt time.Time) *Object {
	return &Object{
		id:         id,
		name:       name,
		objectType: objectType,
		properties: properties,
		tags:       tags,
		createdAt:  createdAt,
	}
}

func (o *Object) ID() ObjectID {
	return o.id
}

func (o *Object) Name() ObjectName {
	return o.name
}

func (o *Object) ObjectType() ObjectType {
	return o.objectType
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

func (o *Object) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Object) UpdateName(name ObjectName) *Object {
	return &Object{
		id:         o.id,
		name:       name,
		objectType: o.objectType,
		properties: o.properties,
		tags:       o.tags,
		createdAt:  o.createdAt,
	}
}

func (o *Object) UpdateProperties(properties map[string]interface{}) *Object {
	return &Object{
		id:         o.id,
		name:       o.name,
		objectType: o.objectType,
		properties: properties,
		tags:       o.tags,
		createdAt:  o.createdAt,
	}
}

func (o *Object) UpdateTags(tags []string) *Object {
	return &Object{
		id:         o.id,
		name:       o.name,
		objectType: o.objectType,
		properties: o.properties,
		tags:       tags,
		createdAt:  o.createdAt,
	}
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

func (o *Object) Equals(other *Object) bool {
	if other == nil {
		return false
	}
	return o.id.Equals(other.id)
}