package domain

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type Child struct {
	ID               string
	UserID           string
	Name             string
	Timezone         string    // IANA timezone, e.g. "Europe/Warsaw" — used to convert UTC→local for clock resolution
	AvatarPath       string
	DefaultProfileID string
	ClockToken       string    // 64-char hex, immutable, used as public auth token for the clock view
	Version          int
	CreatedAt        time.Time // UTC
	UpdatedAt        time.Time // UTC
}

func NewChild(userID, name, timezone string) (*Child, error) {
	if name == "" {
		return nil, ErrEmptyName
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return nil, ErrInvalidTimezone
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	token, err := newClockToken()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &Child{
		ID:         id.String(),
		UserID:     userID,
		Name:       name,
		Timezone:   timezone,
		ClockToken: token,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func newClockToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
