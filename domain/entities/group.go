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

// Group represents an Authentik group for sharing collections
// This is a lightweight entity mainly for authentication/authorization
type Group struct {
	id        GroupID
	name      GroupName
	createdAt time.Time
	updatedAt time.Time
}

type GroupProps struct {
	ID   GroupID
	Name GroupName
}

func NewGroup(props GroupProps) (*Group, error) {
	now := time.Now()
	return &Group{
		id:        props.ID,
		name:      props.Name,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func ReconstructGroup(id GroupID, name GroupName, createdAt, updatedAt time.Time) *Group {
	return &Group{
		id:        id,
		name:      name,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (g *Group) ID() GroupID {
	return g.id
}

func (g *Group) Name() GroupName {
	return g.name
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

func (g *Group) Equals(other *Group) bool {
	if other == nil {
		return false
	}
	return g.id.Equals(other.id)
}