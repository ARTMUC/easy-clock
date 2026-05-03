package userpersistence

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	domainuser "easy-clock/internal/domain/user"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Save(ctx context.Context, rt *domainuser.RefreshToken) error {
	m := RefreshTokenModel{
		ID:        rt.ID,
		UserID:    rt.UserID,
		TokenHash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt,
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) FindByHash(ctx context.Context, hash string) (*domainuser.RefreshToken, error) {
	var m RefreshTokenModel
	err := r.db.WithContext(ctx).Where("token_hash = ?", hash).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainuser.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("find refresh token: %w", err)
	}
	return &domainuser.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt,
	}, nil
}

func (r *RefreshTokenRepository) Delete(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&RefreshTokenModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}
	return nil
}

func (r *RefreshTokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&RefreshTokenModel{}).Error; err != nil {
		return fmt.Errorf("delete all refresh tokens: %w", err)
	}
	return nil
}
