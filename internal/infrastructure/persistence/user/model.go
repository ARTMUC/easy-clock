package userpersistence

import (
	"time"

	"gorm.io/gorm"
)

// UserWriteModel contains only business/domain fields used for INSERT and UPDATE.
// Timestamp columns (created_at, updated_at, deleted_at) are intentionally absent
// so that MySQL can manage them via DEFAULT CURRENT_TIMESTAMP / ON UPDATE CURRENT_TIMESTAMP.
type UserWriteModel struct {
	ID                string `gorm:"primaryKey;type:char(36)"`
	Name              string
	Email             string `gorm:"uniqueIndex;type:varchar(255)"`
	PasswordHash      string `gorm:"type:varchar(255)"`
	Active            bool   `gorm:"type:tinyint(1);default:0"`
	VerificationToken string `gorm:"type:varchar(64)"`
}

func (UserWriteModel) TableName() string { return "users" }

// UserModel is the read model. It embeds UserWriteModel and adds the timestamp
// fields populated by GORM when reading rows from the database.
type UserModel struct {
	UserWriteModel
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}
