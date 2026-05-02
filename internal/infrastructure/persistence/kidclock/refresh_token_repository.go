package kidclock

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"starter/internal/domain"
)

type refreshTokenModel struct {
	ID        string    `gorm:"column:id;primaryKey"`
	UserID    string    `gorm:"column:user_id"`
	TokenHash string    `gorm:"column:token_hash"`
	ExpiresAt time.Time `gorm:"column:expires_at"` // UTC
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (refreshTokenModel) TableName() string { return "refresh_tokens" }

type refreshTokenRepository struct{ db *gorm.DB }

func NewRefreshTokenRepository(db *gorm.DB) domain.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Save(ctx context.Context, t *domain.RefreshToken) error {
	m := refreshTokenModel{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt,
		CreatedAt: t.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *refreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var m refreshTokenModel
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domain.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
	}, nil
}

func (r *refreshTokenRepository) DeleteByHash(ctx context.Context, hash string) error {
	return r.db.WithContext(ctx).Where("token_hash = ?", hash).Delete(&refreshTokenModel{}).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now().UTC()).Delete(&refreshTokenModel{}).Error
}
