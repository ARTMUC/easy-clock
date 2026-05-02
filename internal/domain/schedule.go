package domain

import (
	"time"

	"github.com/google/uuid"
)

// DayAssignment maps one day of the week to a profile for a given child.
// DayOfWeek: 0=Sunday, 1=Monday, ..., 6=Saturday (matches time.Weekday).
// Lookup is done against the child's LOCAL day — convert UTC→local before querying.
type DayAssignment struct {
	ID        string
	ChildID   string
	DayOfWeek int       // local weekday in child's timezone
	ProfileID string
	Version   int
	CreatedAt time.Time // UTC
	UpdatedAt time.Time // UTC
}

func NewDayAssignment(childID string, dayOfWeek int, profileID string) (*DayAssignment, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &DayAssignment{
		ID:        id.String(),
		ChildID:   childID,
		DayOfWeek: dayOfWeek,
		ProfileID: profileID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
