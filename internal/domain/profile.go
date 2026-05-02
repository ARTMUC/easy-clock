package domain

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID         string
	ChildID    string
	Name       string
	Color      string
	Activities []Activity
	Version    int
	CreatedAt  time.Time // UTC
	UpdatedAt  time.Time // UTC
}

type Activity struct {
	ID        string
	ProfileID string
	PresetID  string
	Emoji     string
	Label     string
	FromHour  int       // LOCAL hour in child's timezone (0–23)
	ToHour    int       // LOCAL hour in child's timezone (0–23), exclusive
	Ring      int       // 1=AM, 2=PM
	ImagePath string
	SortOrder int
	Version   int
	CreatedAt time.Time // UTC
	UpdatedAt time.Time // UTC
}

func NewProfile(childID, name, color string) (*Profile, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if color == "" {
		color = "#e07a3a"
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Profile{
		ID:        id.String(),
		ChildID:   childID,
		Name:      name,
		Color:     color,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func NewActivity(profileID, presetID, emoji, label, imagePath string, fromHour, toHour, ring, sortOrder int) (*Activity, error) {
	if fromHour >= toHour {
		return nil, ErrInvalidHourRange
	}
	if imagePath == "" {
		return nil, ErrImageRequired
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Activity{
		ID:        id.String(),
		ProfileID: profileID,
		PresetID:  presetID,
		Emoji:     emoji,
		Label:     label,
		FromHour:  fromHour,
		ToHour:    toHour,
		Ring:      ring,
		ImagePath: imagePath,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// AddActivity appends an activity to the profile after validating ring overlap.
func (p *Profile) AddActivity(a Activity) error {
	if a.FromHour >= a.ToHour {
		return ErrInvalidHourRange
	}
	if a.ImagePath == "" {
		return ErrImageRequired
	}
	for _, existing := range p.Activities {
		if existing.Ring == a.Ring && hoursOverlap(existing.FromHour, existing.ToHour, a.FromHour, a.ToHour) {
			return ErrActivityOverlap
		}
	}
	p.Activities = append(p.Activities, a)
	return nil
}

func hoursOverlap(aFrom, aTo, bFrom, bTo int) bool {
	return aFrom < bTo && bFrom < aTo
}
