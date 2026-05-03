package userpersistence

import "time"

type RefreshTokenModel struct {
	ID        string    `gorm:"primaryKey;type:char(36)"`
	UserID    string    `gorm:"type:char(36);index:idx_rt_user_id"`
	TokenHash string    `gorm:"type:char(64);uniqueIndex"`
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }
