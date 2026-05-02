package userpersistence

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	domainuser "starter/internal/domain/user"
	"starter/internal/repository"
)

// UserRepository is the GORM-backed adapter for domainuser.Repository.
type UserRepository struct {
	repository.BaseRepo[UserModel]
	translator Translator
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		BaseRepo:   repository.NewBaseRepo[UserModel](db),
		translator: Translator{},
	}
}

// Save persists the User inside a transaction.
// INSERT uses tx.Create on the WriteModel so MySQL sets created_at/updated_at via defaults.
// UPDATE uses tx.Save on the WriteModel so no zero time.Time fields are sent.
func (r *UserRepository) Save(ctx context.Context, u *domainuser.User) error {
	wm := r.translator.ToModel(u)

	return r.Transaction(ctx, func(tx *gorm.DB) error {
		if u.ID() == "" {
			if err := tx.Create(wm).Error; err != nil {
				return fmt.Errorf("create user: %w", err)
			}
			u.SetID(wm.ID)
		} else {
			if err := tx.Save(wm).Error; err != nil {
				return fmt.Errorf("update user: %w", err)
			}
		}
		return nil
	})
}

// FindByVerificationToken loads a User by their verification token.
func (r *UserRepository) FindByVerificationToken(ctx context.Context, token string) (*domainuser.User, error) {
	scope := r.DB().WithContext(ctx).Where("verification_token = ?", token)

	m, err := r.FirstScoped(scope)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainuser.ErrInvalidToken
		}
		return nil, fmt.Errorf("find user by token: %w", err)
	}
	return r.translator.ToDomain(m), nil
}

// FindByEmail loads a User by their email address.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	scope := r.DB().WithContext(ctx).Where("email = ?", email)

	m, err := r.FirstScoped(scope)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainuser.ErrNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return r.translator.ToDomain(m), nil
}
