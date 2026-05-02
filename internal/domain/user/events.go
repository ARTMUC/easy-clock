package user

import "time"

// DomainEvent is the base interface for all User domain events.
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// -------------------------------------------------------
// UserCreated
// -------------------------------------------------------

// UserCreated is raised when a new User is registered.
type UserCreated struct {
	UserID     string
	Name       string
	OccurredOn time.Time
}

func (e UserCreated) EventName() string     { return "user.created" }
func (e UserCreated) OccurredAt() time.Time { return e.OccurredOn }
