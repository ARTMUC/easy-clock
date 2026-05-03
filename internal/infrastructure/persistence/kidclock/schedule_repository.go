package kidclock

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"easy-clock/internal/domain"
)

type scheduleDayModel struct {
	ID        string         `gorm:"column:id;primaryKey"`
	ChildID   string         `gorm:"column:child_id"`
	DayOfWeek int            `gorm:"column:day_of_week"` // local weekday in child's timezone
	ProfileID string         `gorm:"column:profile_id"`
	Version   int            `gorm:"column:version"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (scheduleDayModel) TableName() string { return "schedule_days" }

type scheduleRepository struct{ db *gorm.DB }

func NewScheduleRepository(db *gorm.DB) domain.ScheduleRepository {
	return &scheduleRepository{db: db}
}

func (r *scheduleRepository) FindByChildID(ctx context.Context, childID string) ([]domain.DayAssignment, error) {
	var ms []scheduleDayModel
	err := r.db.WithContext(ctx).Where("child_id = ?", childID).Order("day_of_week ASC").Find(&ms).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.DayAssignment, len(ms))
	for i, m := range ms {
		out[i] = scheduleToDomain(&m)
	}
	return out, nil
}

func (r *scheduleRepository) FindDay(ctx context.Context, childID string, dayOfWeek int) (*domain.DayAssignment, error) {
	var m scheduleDayModel
	err := r.db.WithContext(ctx).
		Where("child_id = ? AND day_of_week = ?", childID, dayOfWeek).
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	a := scheduleToDomain(&m)
	return &a, nil
}

func (r *scheduleRepository) Upsert(ctx context.Context, a *domain.DayAssignment) error {
	m := scheduleFromDomain(a)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "child_id"}, {Name: "day_of_week"}},
		DoUpdates: clause.AssignmentColumns([]string{"profile_id", "version", "updated_at", "deleted_at"}),
	}).Create(&m).Error
}

func (r *scheduleRepository) UpsertMany(ctx context.Context, assignments []domain.DayAssignment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range assignments {
			m := scheduleFromDomain(&assignments[i])
			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "child_id"}, {Name: "day_of_week"}},
				DoUpdates: clause.AssignmentColumns([]string{"profile_id", "version", "updated_at", "deleted_at"}),
			}).Create(&m).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *scheduleRepository) DeleteDay(ctx context.Context, childID string, dayOfWeek int) error {
	return r.db.WithContext(ctx).
		Where("child_id = ? AND day_of_week = ?", childID, dayOfWeek).
		Delete(&scheduleDayModel{}).Error
}

func scheduleToDomain(m *scheduleDayModel) domain.DayAssignment {
	return domain.DayAssignment{
		ID:        m.ID,
		ChildID:   m.ChildID,
		DayOfWeek: m.DayOfWeek,
		ProfileID: m.ProfileID,
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func scheduleFromDomain(a *domain.DayAssignment) scheduleDayModel {
	return scheduleDayModel{
		ID:        a.ID,
		ChildID:   a.ChildID,
		DayOfWeek: a.DayOfWeek,
		ProfileID: a.ProfileID,
		Version:   a.Version,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
