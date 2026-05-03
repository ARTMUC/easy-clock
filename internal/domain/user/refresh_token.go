package user

import "time"

const RefreshTTL = 30 * 24 * time.Hour

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
}
