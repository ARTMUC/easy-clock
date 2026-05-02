package kidclock

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"starter/internal/domain"
)

type profileModel struct {
	ID         string          `gorm:"column:id;primaryKey"`
	ChildID    string          `gorm:"column:child_id"`
	Name       string          `gorm:"column:name"`
	Color      string          `gorm:"column:color"`
	Version    int             `gorm:"column:version"`
	CreatedAt  time.Time       `gorm:"column:created_at"`
	UpdatedAt  time.Time       `gorm:"column:updated_at"`
	DeletedAt  gorm.DeletedAt  `gorm:"column:deleted_at;index"`
	Activities []activityModel `gorm:"foreignKey:ProfileID"`
}

func (profileModel) TableName() string { return "profiles" }

type profileRepository struct{ db *gorm.DB }

func NewProfileRepository(db *gorm.DB) domain.ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) FindByID(ctx context.Context, id string) (*domain.Profile, error) {
	var m profileModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return profileToDomain(&m), err
}

func (r *profileRepository) FindWithActivities(ctx context.Context, id string) (*domain.Profile, error) {
	var m profileModel
	err := r.db.WithContext(ctx).
		Preload("Activities", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return profileToDomain(&m), err
}

func (r *profileRepository) FindByChildID(ctx context.Context, childID string) ([]domain.Profile, error) {
	var ms []profileModel
	err := r.db.WithContext(ctx).Where("child_id = ?", childID).Find(&ms).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Profile, len(ms))
	for i, m := range ms {
		out[i] = *profileToDomain(&m)
	}
	return out, nil
}

func (r *profileRepository) Save(ctx context.Context, p *domain.Profile) error {
	m := profileFromDomain(p)
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *profileRepository) Update(ctx context.Context, p *domain.Profile) error {
	m := profileFromDomain(p)
	m.Version++
	m.UpdatedAt = time.Now().UTC()
	res := r.db.WithContext(ctx).Where("id = ? AND version = ?", p.ID, p.Version).Save(&m)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrOptimisticLock
	}
	p.Version = m.Version
	return nil
}

func (r *profileRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&profileModel{}).Error
}

func profileToDomain(m *profileModel) *domain.Profile {
	p := &domain.Profile{
		ID:        m.ID,
		ChildID:   m.ChildID,
		Name:      m.Name,
		Color:     m.Color,
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	for _, a := range m.Activities {
		p.Activities = append(p.Activities, *activityToDomain(&a))
	}
	return p
}

func profileFromDomain(p *domain.Profile) profileModel {
	return profileModel{
		ID:        p.ID,
		ChildID:   p.ChildID,
		Name:      p.Name,
		Color:     p.Color,
		Version:   p.Version,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
