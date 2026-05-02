package user

import (
	"time"

	"github.com/google/uuid"
)

// User is the aggregate root for the User bounded context.
type User struct {
	id                string
	name              string
	email             string
	passwordHash      string
	active            bool
	verificationToken string

	events []DomainEvent
}

// -----------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------

// NewUser creates a bare User (no auth — legacy path).
func NewUser(name string) (*User, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	u := &User{name: name}
	u.record(UserCreated{Name: name, OccurredOn: time.Now()})
	return u, nil
}

// NewUserWithAuth creates a new User with email/password and a verification token.
// active is false until VerifyEmail is called.
func NewUserWithAuth(name, email, passwordHash, verificationToken string) (*User, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if email == "" {
		return nil, ErrEmptyEmail
	}
	u := &User{
		id:                uuid.New().String(),
		name:              name,
		email:             email,
		passwordHash:      passwordHash,
		active:            false,
		verificationToken: verificationToken,
	}
	u.record(UserCreated{Name: name, OccurredOn: time.Now()})
	return u, nil
}

// Reconstitute rebuilds a User from persistence data without triggering events.
func Reconstitute(id string, name, email, passwordHash string, active bool, verificationToken string) *User {
	return &User{
		id:                id,
		name:              name,
		email:             email,
		passwordHash:      passwordHash,
		active:            active,
		verificationToken: verificationToken,
	}
}

// SetID is called by the repository after the DB assigns an ID.
func (u *User) SetID(id string) { u.id = id }

// -----------------------------------------------------------------
// Queries
// -----------------------------------------------------------------

func (u *User) ID() string              { return u.id }
func (u *User) Name() string            { return u.name }
func (u *User) Email() string           { return u.email }
func (u *User) PasswordHash() string    { return u.passwordHash }
func (u *User) Active() bool            { return u.active }
func (u *User) VerificationToken() string { return u.verificationToken }

// -----------------------------------------------------------------
// Commands
// -----------------------------------------------------------------

// Activate marks the user as active and clears the verification token.
func (u *User) Activate() {
	u.active = true
	u.verificationToken = ""
}

func (u *User) Rename(name string) error {
	if name == "" {
		return ErrEmptyName
	}
	u.name = name
	return nil
}

// -----------------------------------------------------------------
// Domain events
// -----------------------------------------------------------------

func (u *User) PullEvents() []DomainEvent {
	events := u.events
	u.events = nil
	return events
}

func (u *User) record(e DomainEvent) {
	u.events = append(u.events, e)
}
