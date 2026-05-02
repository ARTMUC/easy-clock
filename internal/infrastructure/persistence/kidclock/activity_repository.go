package kidclock

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"starter/internal/domain"
)

type activityModel struct {
	ID        string         `gorm:"column:id;primaryKey"`
	ProfileID string         `gorm:"column:profile_id"`
	PresetID  *string        `gorm:"column:preset_id"`
	Emoji     string         `gorm:"column:emoji"`
	Label     string         `gorm:"column:label"`
	FromHour  int            `gorm:"column:from_hour"` // LOCAL hour in child's timezone
	ToHour    int            `gorm:"column:to_hour"`   // LOCAL hour in child's timezone
	Ring      int            `gorm:"column:ring"`
	ImagePath string         `gorm:"column:image_path"`
	SortOrder int            `gorm:"column:sort_order"`
	Version   int            `gorm:"column:version"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (activityModel) TableName() string { return "activities" }

type activityRepository struct{ db *gorm.DB }

func NewActivityRepository(db *gorm.DB) domain.ActivityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) FindByID(ctx context.Context, id string) (*domain.Activity, error) {
	var m activityModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return activityToDomain(&m), err
}

func (r *activityRepository) Save(ctx context.Context, a *domain.Activity) error {
	m := activityFromDomain(a)
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *activityRepository) Update(ctx context.Context, a *domain.Activity) error {
	m := activityFromDomain(a)
	m.Version++
	m.UpdatedAt = time.Now().UTC()
	res := r.db.WithContext(ctx).Where("id = ? AND version = ?", a.ID, a.Version).Save(&m)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrOptimisticLock
	}
	a.Version = m.Version
	return nil
}

func (r *activityRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&activityModel{}).Error
}

func activityToDomain(m *activityModel) *domain.Activity {
	a := &domain.Activity{
		ID:        m.ID,
		ProfileID: m.ProfileID,
		Emoji:     m.Emoji,
		Label:     m.Label,
		FromHour:  m.FromHour,
		ToHour:    m.ToHour,
		Ring:      m.Ring,
		ImagePath: m.ImagePath,
		SortOrder: m.SortOrder,
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.PresetID != nil {
		a.PresetID = *m.PresetID
	}
	return a
}

func activityFromDomain(a *domain.Activity) activityModel {
	return activityModel{
		ID:        a.ID,
		ProfileID: a.ProfileID,
		PresetID:  nilStr(a.PresetID),
		Emoji:     a.Emoji,
		Label:     a.Label,
		FromHour:  a.FromHour,
		ToHour:    a.ToHour,
		Ring:      a.Ring,
		ImagePath: a.ImagePath,
		SortOrder: a.SortOrder,
		Version:   a.Version,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
