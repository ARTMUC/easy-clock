package domain

import (
	"time"

	"github.com/google/uuid"
)

type PresetActivity struct {
	ID        string
	Emoji     string
	Label     string
	ImagePath string
	Ring      int
	SortOrder int
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPresetActivity(emoji, label, imagePath string, ring, sortOrder int) (*PresetActivity, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &PresetActivity{
		ID:        id.String(),
		Emoji:     emoji,
		Label:     label,
		ImagePath: imagePath,
		Ring:      ring,
		SortOrder: sortOrder,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
