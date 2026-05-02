package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Save(ctx context.Context, u *User) error
	// Update saves only mutable fields; returns ErrOptimisticLock on version mismatch.
	Update(ctx context.Context, u *User) error
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, t *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	DeleteByHash(ctx context.Context, hash string) error
	DeleteExpired(ctx context.Context) error
}

type ChildRepository interface {
	FindByID(ctx context.Context, id string) (*Child, error)
	FindByUserID(ctx context.Context, userID string) ([]Child, error)
	FindByClockToken(ctx context.Context, token string) (*Child, error)
	Save(ctx context.Context, c *Child) error
	// Update saves mutable fields (name, timezone, avatar_path, default_profile_id);
	// returns ErrOptimisticLock on version mismatch.
	Update(ctx context.Context, c *Child) error
	Delete(ctx context.Context, id string) error
}

type ProfileRepository interface {
	FindByID(ctx context.Context, id string) (*Profile, error)
	// FindWithActivities loads the profile and all its activities in one call.
	FindWithActivities(ctx context.Context, id string) (*Profile, error)
	FindByChildID(ctx context.Context, childID string) ([]Profile, error)
	Save(ctx context.Context, p *Profile) error
	Update(ctx context.Context, p *Profile) error
	Delete(ctx context.Context, id string) error
}

type ActivityRepository interface {
	FindByID(ctx context.Context, id string) (*Activity, error)
	Save(ctx context.Context, a *Activity) error
	Update(ctx context.Context, a *Activity) error
	Delete(ctx context.Context, id string) error
}

type PresetActivityRepository interface {
	FindAll(ctx context.Context) ([]PresetActivity, error)
	FindByID(ctx context.Context, id string) (*PresetActivity, error)
}

type ScheduleRepository interface {
	// FindByChildID returns all DayAssignments for the child (0–7 entries).
	FindByChildID(ctx context.Context, childID string) ([]DayAssignment, error)
	// FindDay returns the assignment for a specific local weekday, or ErrNotFound.
	FindDay(ctx context.Context, childID string, dayOfWeek int) (*DayAssignment, error)
	// Upsert inserts or replaces the assignment for child+day.
	Upsert(ctx context.Context, a *DayAssignment) error
	// UpsertMany bulk-upserts multiple days in a single transaction.
	UpsertMany(ctx context.Context, assignments []DayAssignment) error
	DeleteDay(ctx context.Context, childID string, dayOfWeek int) error
}

type EventRepository interface {
	FindByID(ctx context.Context, id string) (*Event, error)
	// FindByChildID returns events in [from, to] date range (LOCAL dates).
	FindByChildID(ctx context.Context, childID string, from, to time.Time) ([]Event, error)
	// FindForDate returns events whose Date matches the given LOCAL date string "2006-01-02".
	FindForDate(ctx context.Context, childID string, date string) ([]Event, error)
	Save(ctx context.Context, e *Event) error
	Update(ctx context.Context, e *Event) error
	Delete(ctx context.Context, id string) error
}
