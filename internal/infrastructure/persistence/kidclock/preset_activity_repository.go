package kidclock

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"easy-clock/internal/domain"
)

type presetActivityModel struct {
	ID        string         `gorm:"column:id;primaryKey"`
	Emoji     string         `gorm:"column:emoji"`
	Label     string         `gorm:"column:label"`
	ImagePath string         `gorm:"column:image_path"`
	SortOrder int            `gorm:"column:sort_order"`
	Version   int            `gorm:"column:version"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (presetActivityModel) TableName() string { return "preset_activities" }

type presetActivityRepository struct{ db *gorm.DB }

func NewPresetActivityRepository(db *gorm.DB) domain.PresetActivityRepository {
	return &presetActivityRepository{db: db}
}

func (r *presetActivityRepository) FindAll(ctx context.Context) ([]domain.PresetActivity, error) {
	var ms []presetActivityModel
	err := r.db.WithContext(ctx).Order("sort_order ASC").Find(&ms).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.PresetActivity, len(ms))
	for i, m := range ms {
		out[i] = presetToDomain(&m)
	}
	return out, nil
}

func (r *presetActivityRepository) FindByID(ctx context.Context, id string) (*domain.PresetActivity, error) {
	var m presetActivityModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p := presetToDomain(&m)
	return &p, nil
}

func presetToDomain(m *presetActivityModel) domain.PresetActivity {
	return domain.PresetActivity{
		ID:        m.ID,
		Emoji:     m.Emoji,
		Label:     m.Label,
		ImagePath: m.ImagePath,
		SortOrder: m.SortOrder,
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
