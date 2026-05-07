package kidclock

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"easy-clock/internal/domain"
)

type eventModel struct {
	ID         string               `gorm:"column:id;primaryKey"`
	ChildID    string               `gorm:"column:child_id"`
	Date       string               `gorm:"column:date"`      // LOCAL date "2006-01-02" in child's timezone
	FromTime   string               `gorm:"column:from_time"` // LOCAL time "HH:MM:SS" in child's timezone
	ToTime     string               `gorm:"column:to_time"`   // LOCAL time "HH:MM:SS" in child's timezone
	Label      string               `gorm:"column:label"`
	Emoji      *string              `gorm:"column:emoji"`
	ImagePath  *string              `gorm:"column:image_path"`
	ProfileID  *string              `gorm:"column:profile_id"`
	Version    int                  `gorm:"column:version"`
	CreatedAt  time.Time            `gorm:"column:created_at"` // UTC
	UpdatedAt  time.Time            `gorm:"column:updated_at"` // UTC
	DeletedAt  gorm.DeletedAt       `gorm:"column:deleted_at;index"`
	Activities []eventActivityModel `gorm:"foreignKey:EventID"`
}

func (eventModel) TableName() string { return "events" }

type eventActivityModel struct {
	ID        string         `gorm:"column:id;primaryKey"`
	EventID   string         `gorm:"column:event_id"`
	Emoji     string         `gorm:"column:emoji"`
	Label     string         `gorm:"column:label"`
	FromHour  int            `gorm:"column:from_hour"` // LOCAL hour in child's timezone
	ToHour    int            `gorm:"column:to_hour"`   // LOCAL hour in child's timezone
	ImagePath *string        `gorm:"column:image_path"`
	Version   int            `gorm:"column:version"`
	CreatedAt time.Time      `gorm:"column:created_at"` // UTC
	UpdatedAt time.Time      `gorm:"column:updated_at"` // UTC
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (eventActivityModel) TableName() string { return "event_activities" }

type eventRepository struct{ db *gorm.DB }

func NewEventRepository(db *gorm.DB) domain.EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) FindByID(ctx context.Context, id string) (*domain.Event, error) {
	var m eventModel
	err := r.db.WithContext(ctx).Preload("Activities").Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	return eventToDomain(&m), err
}

func (r *eventRepository) FindByChildID(ctx context.Context, childID string, from, to time.Time) ([]domain.Event, error) {
	var ms []eventModel
	err := r.db.WithContext(ctx).Preload("Activities").
		Where("child_id = ? AND date >= ? AND date <= ?",
			childID,
			from.Format("2006-01-02"),  // LOCAL date boundaries
			to.Format("2006-01-02"),
		).
		Order("date ASC, from_time ASC").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Event, len(ms))
	for i, m := range ms {
		out[i] = *eventToDomain(&m)
	}
	return out, nil
}

func (r *eventRepository) FindForDate(ctx context.Context, childID string, date string) ([]domain.Event, error) {
	var ms []eventModel
	err := r.db.WithContext(ctx).Preload("Activities").
		Where("child_id = ? AND date = ?", childID, date).
		Order("from_time ASC").
		Find(&ms).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Event, len(ms))
	for i, m := range ms {
		out[i] = *eventToDomain(&m)
	}
	return out, nil
}

func (r *eventRepository) Save(ctx context.Context, e *domain.Event) error {
	m := eventFromDomain(e)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		for _, a := range m.Activities {
			if err := tx.Create(&a).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *eventRepository) Update(ctx context.Context, e *domain.Event) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		m := eventFromDomain(e)
		m.Version++
		m.UpdatedAt = time.Now().UTC()
		res := tx.Where("id = ? AND version = ?", e.ID, e.Version).Save(&m)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrOptimisticLock
		}
		// Replace inline activities: delete old, insert new.
		if err := tx.Where("event_id = ?", e.ID).Delete(&eventActivityModel{}).Error; err != nil {
			return err
		}
		for _, a := range e.Activities {
			am := eventActivityFromDomain(&a)
			if err := tx.Create(&am).Error; err != nil {
				return err
			}
		}
		e.Version++
		return nil
	})
}

func (r *eventRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&eventModel{}).Error
}

func eventToDomain(m *eventModel) *domain.Event {
	date, _ := time.Parse("2006-01-02", m.Date)
	e := &domain.Event{
		ID:        m.ID,
		ChildID:   m.ChildID,
		Date:      date,
		FromTime:  m.FromTime,
		ToTime:    m.ToTime,
		Label:     m.Label,
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.Emoji != nil {
		e.Emoji = *m.Emoji
	}
	if m.ImagePath != nil {
		e.ImagePath = *m.ImagePath
	}
	if m.ProfileID != nil {
		e.ProfileID = *m.ProfileID
	}
	for _, a := range m.Activities {
		e.Activities = append(e.Activities, eventActivityToDomain(&a))
	}
	return e
}

func eventFromDomain(e *domain.Event) eventModel {
	m := eventModel{
		ID:        e.ID,
		ChildID:   e.ChildID,
		Date:      e.Date.Format("2006-01-02"),
		FromTime:  e.FromTime,
		ToTime:    e.ToTime,
		Label:     e.Label,
		Emoji:     nilStr(e.Emoji),
		ImagePath: nilStr(e.ImagePath),
		ProfileID: nilStr(e.ProfileID),
		Version:   e.Version,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
	for _, a := range e.Activities {
		m.Activities = append(m.Activities, eventActivityFromDomain(&a))
	}
	return m
}

func eventActivityToDomain(m *eventActivityModel) domain.EventActivity {
	a := domain.EventActivity{
		ID:        m.ID,
		EventID:   m.EventID,
		Emoji:     m.Emoji,
		Label:     m.Label,
		FromHour:  m.FromHour,
		ToHour:    m.ToHour,
		Version:   m.Version,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.ImagePath != nil {
		a.ImagePath = *m.ImagePath
	}
	return a
}

func eventActivityFromDomain(a *domain.EventActivity) eventActivityModel {
	return eventActivityModel{
		ID:        a.ID,
		EventID:   a.EventID,
		Emoji:     a.Emoji,
		Label:     a.Label,
		FromHour:  a.FromHour,
		ToHour:    a.ToHour,
		ImagePath: nilStr(a.ImagePath),
		Version:   a.Version,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}
