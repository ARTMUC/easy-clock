package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID         string
	ChildID    string
	Date       time.Time // LOCAL date in child's timezone — stored as DATE, no time component
	FromTime   string    // LOCAL time in child's timezone, "HH:MM:SS"
	ToTime     string    // LOCAL time in child's timezone, "HH:MM:SS"
	Label      string
	Emoji      string
	ImagePath  string
	ProfileID  string
	Activities []EventActivity
	Version    int
	CreatedAt  time.Time // UTC
	UpdatedAt  time.Time // UTC
}

type EventActivity struct {
	ID        string
	EventID   string
	Emoji     string
	Label     string
	FromHour  int       // LOCAL hour in child's timezone (0–23)
	ToHour    int       // LOCAL hour in child's timezone (0–23), exclusive
	ImagePath string
	Version   int
	CreatedAt time.Time // UTC
	UpdatedAt time.Time // UTC
}

func NewEvent(childID string, date time.Time, fromTime, toTime, label string) (*Event, error) {
	if fromTime >= toTime {
		return nil, ErrInvalidTimeRange
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Event{
		ID:        id.String(),
		ChildID:   childID,
		Date:      date,
		FromTime:  fromTime,
		ToTime:    toTime,
		Label:     label,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (e *Event) SetProfile(profileID string) error {
	if len(e.Activities) > 0 {
		return ErrEventProfileXorActivities
	}
	e.ProfileID = profileID
	return nil
}

func (e *Event) AddActivity(a EventActivity) error {
	if e.ProfileID != "" {
		return ErrEventProfileXorActivities
	}
	if a.FromHour >= a.ToHour {
		return ErrInvalidHourRange
	}
	e.Activities = append(e.Activities, a)
	return nil
}
