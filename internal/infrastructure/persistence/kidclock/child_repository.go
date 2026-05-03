package kidclock

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"easy-clock/internal/domain"
)

type childModel struct {
	ID               string         `gorm:"column:id;primaryKey"`
	UserID           string         `gorm:"column:user_id"`
	Name             string         `gorm:"column:name"`
	Timezone         string         `gorm:"column:timezone"`
	AvatarPath       *string        `gorm:"column:avatar_path"`
	DefaultProfileID *string        `gorm:"column:default_profile_id"`
	ClockToken       string         `gorm:"column:clock_token"`
	Version          int            `gorm:"column:version"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (childModel) TableName() string { return "children" }

type childRepository struct{ db *gorm.DB }

func NewChildRepository(db *gorm.DB) domain.ChildRepository {
	return &childRepository{db: db}
}

func (r *childRepository) FindByID(ctx context.Context, id string) (*domain.Child, error) {
	var m childModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return childToDomain(&m), err
}

func (r *childRepository) FindByUserID(ctx context.Context, userID string) ([]domain.Child, error) {
	var ms []childModel
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&ms).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Child, len(ms))
	for i, m := range ms {
		out[i] = *childToDomain(&m)
	}
	return out, nil
}

func (r *childRepository) FindByClockToken(ctx context.Context, token string) (*domain.Child, error) {
	var m childModel
	err := r.db.WithContext(ctx).Where("clock_token = ?", token).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return childToDomain(&m), err
}

func (r *childRepository) Save(ctx context.Context, c *domain.Child) error {
	m := childFromDomain(c)
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *childRepository) Update(ctx context.Context, c *domain.Child) error {
	m := childFromDomain(c)
	m.Version++
	m.UpdatedAt = time.Now().UTC()
	res := r.db.WithContext(ctx).Where("id = ? AND version = ?", c.ID, c.Version).Save(&m)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrOptimisticLock
	}
	c.Version = m.Version
	return nil
}

func (r *childRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&childModel{}).Error
}

func childToDomain(m *childModel) *domain.Child {
	c := &domain.Child{
		ID:         m.ID,
		UserID:     m.UserID,
		Name:       m.Name,
		Timezone:   m.Timezone,
		ClockToken: m.ClockToken,
		Version:    m.Version,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
	if m.AvatarPath != nil {
		c.AvatarPath = *m.AvatarPath
	}
	if m.DefaultProfileID != nil {
		c.DefaultProfileID = *m.DefaultProfileID
	}
	return c
}

func childFromDomain(c *domain.Child) childModel {
	return childModel{
		ID:               c.ID,
		UserID:           c.UserID,
		Name:             c.Name,
		Timezone:         c.Timezone,
		AvatarPath:       nilStr(c.AvatarPath),
		DefaultProfileID: nilStr(c.DefaultProfileID),
		ClockToken:       c.ClockToken,
		Version:          c.Version,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}
}

// nilStr returns nil for empty string, otherwise a pointer to the value.
func nilStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
