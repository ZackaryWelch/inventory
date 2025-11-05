package entities

import (
	"errors"
	"time"
)

var (
	ErrInvalidGroupID   = errors.New("invalid group ID")
	ErrInvalidGroupName = errors.New("group name must be between 1 and 255 characters")
)

type GroupID struct {
	value string
}

func NewGroupID() GroupID {
	return GroupID{value: ""}
}

func GroupIDFromString(id string) (GroupID, error) {
	if id == "" {
		return GroupID{}, ErrInvalidGroupID
	}
	return GroupID{value: id}, nil
}

func (g GroupID) String() string {
	return g.value
}

func (g GroupID) Equals(other GroupID) bool {
	return g.value == other.value
}

func (g GroupID) IsZero() bool {
	return g.value == ""
}

type GroupName struct {
	value string
}

func NewGroupName(name string) (GroupName, error) {
	if len(name) < 1 || len(name) > 255 {
		return GroupName{}, ErrInvalidGroupName
	}
	return GroupName{value: name}, nil
}

func (g GroupName) String() string {
	return g.value
}

func (g GroupName) Equals(other GroupName) bool {
	return g.value == other.value
}

type GroupDescription struct {
	value string
}

func NewGroupDescription(description string) GroupDescription {
	return GroupDescription{value: description}
}

func (d GroupDescription) String() string {
	return d.value
}

func (d GroupDescription) Equals(other GroupDescription) bool {
	return d.value == other.value
}

// Group represents an Authentik group for sharing collections
// This is a lightweight entity mainly for authentication/authorization
type Group struct {
	id          GroupID
	name        GroupName
	description GroupDescription
	createdAt   time.Time
	updatedAt   time.Time
}

type GroupProps struct {
	ID          GroupID
	Name        GroupName
	Description GroupDescription
}

func NewGroup(props GroupProps) (*Group, error) {
	now := time.Now()
	return &Group{
		id:          props.ID,
		name:        props.Name,
		description: props.Description,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

func ReconstructGroup(id GroupID, name GroupName, description GroupDescription, createdAt, updatedAt time.Time) *Group {
	return &Group{
		id:          id,
		name:        name,
		description: description,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}
}

func (g *Group) ID() GroupID {
	return g.id
}

func (g *Group) Name() GroupName {
	return g.name
}

func (g *Group) Description() GroupDescription {
	return g.description
}

func (g *Group) CreatedAt() time.Time {
	return g.createdAt
}

func (g *Group) UpdatedAt() time.Time {
	return g.updatedAt
}

func (g *Group) UpdateName(name GroupName) error {
	g.name = name
	g.updatedAt = time.Now()
	return nil
}

func (g *Group) UpdateDescription(description GroupDescription) error {
	g.description = description
	g.updatedAt = time.Now()
	return nil
}

func (g *Group) Equals(other *Group) bool {
	if other == nil {
		return false
	}
	return g.id.Equals(other.id)
}
