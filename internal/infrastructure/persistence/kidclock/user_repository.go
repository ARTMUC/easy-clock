package kidclock

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"starter/internal/domain"
)

type userModel struct {
	ID           string         `gorm:"column:id;primaryKey"`
	Email        string         `gorm:"column:email"`
	PasswordHash string         `gorm:"column:password_hash"`
	Version      int            `gorm:"column:version"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (userModel) TableName() string { return "users" }

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return userToDomain(&m), err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m userModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return userToDomain(&m), err
}

func (r *userRepository) Save(ctx context.Context, u *domain.User) error {
	m := userFromDomain(u)
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *userRepository) Update(ctx context.Context, u *domain.User) error {
	m := userFromDomain(u)
	m.Version++
	m.UpdatedAt = time.Now().UTC()
	res := r.db.WithContext(ctx).Where("id = ? AND version = ?", u.ID, u.Version).Save(&m)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrOptimisticLock
	}
	u.Version = m.Version
	return nil
}

func userToDomain(m *userModel) *domain.User {
	return &domain.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		Version:      m.Version,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func userFromDomain(u *domain.User) userModel {
	return userModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Version:      u.Version,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}
