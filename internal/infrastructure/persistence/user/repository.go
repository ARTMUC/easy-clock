package userpersistence

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	domainuser "easy-clock/internal/domain/user"
)

type UserRepository struct {
	db         *gorm.DB
	translator Translator
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db, translator: Translator{}}
}

func (r *UserRepository) Save(ctx context.Context, u *domainuser.User) error {
	wm := r.translator.ToModel(u)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if u.ID() == "" {
			if err := tx.Create(wm).Error; err != nil {
				return fmt.Errorf("create user: %w", err)
			}
		} else {
			if err := tx.Save(wm).Error; err != nil {
				return fmt.Errorf("update user: %w", err)
			}
		}
		return nil
	})
}

func (r *UserRepository) FindByVerificationToken(ctx context.Context, token string) (*domainuser.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).Where("verification_token = ?", token).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainuser.ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("find user by token: %w", err)
	}
	return r.translator.ToDomain(&m), nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	var m UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainuser.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return r.translator.ToDomain(&m), nil
}
