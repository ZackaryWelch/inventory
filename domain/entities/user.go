package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidUsername     = errors.New("username must be between 1 and 255 characters")
	ErrInvalidEmailAddress = errors.New("invalid email address format")
)

type UserID struct {
	value string
}

func NewUserID() UserID {
	return UserID{value: uuid.New().String()}
}

func UserIDFromString(id string) (UserID, error) {
	if id == "" {
		return UserID{}, ErrInvalidUserID
	}
	return UserID{value: id}, nil
}

func (u UserID) String() string {
	return u.value
}

func (u UserID) Equals(other UserID) bool {
	return u.value == other.value
}

type Username struct {
	value string
}

func NewUsername(username string) (Username, error) {
	if len(username) < 1 || len(username) > 255 {
		return Username{}, ErrInvalidUsername
	}
	return Username{value: username}, nil
}

func (u Username) String() string {
	return u.value
}

func (u Username) Equals(other Username) bool {
	return u.value == other.value
}

type EmailAddress struct {
	value string
}

func NewEmailAddress(email string) (EmailAddress, error) {
	if email == "" || len(email) > 255 {
		return EmailAddress{}, ErrInvalidEmailAddress
	}
	// Basic email validation - in production, use a proper email validation library
	if !isValidEmail(email) {
		return EmailAddress{}, ErrInvalidEmailAddress
	}
	return EmailAddress{value: email}, nil
}

func (e EmailAddress) String() string {
	return e.value
}

func (e EmailAddress) Equals(other EmailAddress) bool {
	return e.value == other.value
}

type User struct {
	id           UserID
	username     Username
	emailAddress EmailAddress
	authentikID  string
	createdAt    time.Time
	updatedAt    time.Time
}

type UserProps struct {
	Username     Username
	EmailAddress EmailAddress
	AuthentikID  string
}

func NewUser(props UserProps) (*User, error) {
	now := time.Now()
	return &User{
		id:           NewUserID(),
		username:     props.Username,
		emailAddress: props.EmailAddress,
		authentikID:  props.AuthentikID,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func ReconstructUser(id UserID, username Username, emailAddress EmailAddress, authentikID string, createdAt, updatedAt time.Time) *User {
	return &User{
		id:           id,
		username:     username,
		emailAddress: emailAddress,
		authentikID:  authentikID,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Username() Username {
	return u.username
}

func (u *User) EmailAddress() EmailAddress {
	return u.emailAddress
}

func (u *User) AuthentikID() string {
	return u.authentikID
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) UpdateUsername(username Username) error {
	u.username = username
	u.updatedAt = time.Now()
	return nil
}

func (u *User) UpdateEmailAddress(emailAddress EmailAddress) error {
	u.emailAddress = emailAddress
	u.updatedAt = time.Now()
	return nil
}

func (u *User) Equals(other *User) bool {
	if other == nil {
		return false
	}
	return u.id.Equals(other.id)
}

// Basic email validation - replace with proper validation in production
func isValidEmail(email string) bool {
	// Very basic check - contains @ and .
	atCount := 0
	dotAfterAt := false
	atPos := -1

	for i, char := range email {
		if char == '@' {
			atCount++
			atPos = i
		}
		if char == '.' && atPos != -1 && i > atPos {
			dotAfterAt = true
		}
	}

	return atCount == 1 && dotAfterAt && len(email) > 3
}
