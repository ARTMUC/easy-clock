package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"easy-clock/internal/app"
	"easy-clock/internal/domain"
)

// ---- in-memory stubs ----

type stubChildRepo struct {
	byToken map[string]*domain.Child
	byID    map[string]*domain.Child
}

func (r *stubChildRepo) FindByClockToken(_ context.Context, token string) (*domain.Child, error) {
	if c, ok := r.byToken[token]; ok {
		return c, nil
	}
	return nil, domain.ErrNotFound
}
func (r *stubChildRepo) FindByID(_ context.Context, id string) (*domain.Child, error) {
	if c, ok := r.byID[id]; ok {
		return c, nil
	}
	return nil, domain.ErrNotFound
}
func (r *stubChildRepo) FindByUserID(_ context.Context, _ string) ([]domain.Child, error) { return nil, nil }
func (r *stubChildRepo) Save(_ context.Context, c *domain.Child) error                    { return nil }
func (r *stubChildRepo) Update(_ context.Context, c *domain.Child) error                  { return nil }
func (r *stubChildRepo) Delete(_ context.Context, _ string) error                         { return nil }

type stubEventRepo struct {
	events []domain.Event
}

func (r *stubEventRepo) FindForDate(_ context.Context, childID, date string) ([]domain.Event, error) {
	var out []domain.Event
	for _, e := range r.events {
		if e.ChildID == childID && e.Date.Format("2006-01-02") == date {
			out = append(out, e)
		}
	}
	return out, nil
}
func (r *stubEventRepo) FindByID(_ context.Context, _ string) (*domain.Event, error) { return nil, domain.ErrNotFound }
func (r *stubEventRepo) FindByChildID(_ context.Context, _ string, _, _ time.Time) ([]domain.Event, error) {
	return nil, nil
}
func (r *stubEventRepo) Save(_ context.Context, _ *domain.Event) error   { return nil }
func (r *stubEventRepo) Update(_ context.Context, _ *domain.Event) error { return nil }
func (r *stubEventRepo) Delete(_ context.Context, _ string) error        { return nil }

type stubScheduleRepo struct {
	days map[string]*domain.DayAssignment // key = childID+":"+dayOfWeek
}

func (r *stubScheduleRepo) FindDay(_ context.Context, childID string, dayOfWeek int) (*domain.DayAssignment, error) {
	key := fmt.Sprintf("%s:%d", childID, dayOfWeek)
	if a, ok := r.days[key]; ok {
		return a, nil
	}
	return nil, domain.ErrNotFound
}
func (r *stubScheduleRepo) FindByChildID(_ context.Context, _ string) ([]domain.DayAssignment, error) {
	return nil, nil
}
func (r *stubScheduleRepo) Upsert(_ context.Context, _ *domain.DayAssignment) error { return nil }
func (r *stubScheduleRepo) UpsertMany(_ context.Context, _ []domain.DayAssignment) error {
	return nil
}
func (r *stubScheduleRepo) DeleteDay(_ context.Context, _ string, _ int) error { return nil }

type stubProfileRepo struct {
	profiles map[string]*domain.Profile
}

func (r *stubProfileRepo) FindByID(_ context.Context, id string) (*domain.Profile, error) {
	if p, ok := r.profiles[id]; ok {
		return p, nil
	}
	return nil, domain.ErrNotFound
}
func (r *stubProfileRepo) FindWithActivities(_ context.Context, id string) (*domain.Profile, error) {
	return r.FindByID(context.Background(), id)
}
func (r *stubProfileRepo) FindByChildID(_ context.Context, _ string) ([]domain.Profile, error) {
	return nil, nil
}
func (r *stubProfileRepo) Save(_ context.Context, _ *domain.Profile) error   { return nil }
func (r *stubProfileRepo) Update(_ context.Context, _ *domain.Profile) error { return nil }
func (r *stubProfileRepo) Delete(_ context.Context, _ string) error          { return nil }

// ---- helpers ----

func newSvc(
	children *stubChildRepo,
	events *stubEventRepo,
	schedule *stubScheduleRepo,
	profiles *stubProfileRepo,
) *app.ClockService {
	return app.NewClockService(children, events, schedule, profiles)
}

func scheduleKey(childID string, day int) string {
	return fmt.Sprintf("%s:%d", childID, day)
}

// ---- tests ----

func TestClockService_unknownToken(t *testing.T) {
	svc := newSvc(
		&stubChildRepo{byToken: map[string]*domain.Child{}},
		&stubEventRepo{},
		&stubScheduleRepo{days: map[string]*domain.DayAssignment{}},
		&stubProfileRepo{profiles: map[string]*domain.Profile{}},
	)
	_, err := svc.Resolve(context.Background(), "no-such-token", time.Now())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestClockService_emptyState(t *testing.T) {
	child := &domain.Child{
		ID:         "child-1",
		ClockToken: "tok1",
		Timezone:   "UTC",
	}
	svc := newSvc(
		&stubChildRepo{byToken: map[string]*domain.Child{"tok1": child}},
		&stubEventRepo{},
		&stubScheduleRepo{days: map[string]*domain.DayAssignment{}},
		&stubProfileRepo{profiles: map[string]*domain.Profile{}},
	)
	state, err := svc.Resolve(context.Background(), "tok1", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !state.Empty {
		t.Error("expected Empty state when no profile configured")
	}
}

func TestClockService_defaultProfile(t *testing.T) {
	child := &domain.Child{
		ID:               "child-2",
		ClockToken:       "tok2",
		Timezone:         "UTC",
		DefaultProfileID: "prof-1",
	}
	profile := &domain.Profile{
		ID:   "prof-1",
		Name: "Default",
		Activities: []domain.Activity{
			{ID: "act-1", Emoji: "🌙", Label: "Sleep", FromHour: 22, ToHour: 24, Ring: 2},
		},
	}
	svc := newSvc(
		&stubChildRepo{byToken: map[string]*domain.Child{"tok2": child}},
		&stubEventRepo{},
		&stubScheduleRepo{days: map[string]*domain.DayAssignment{}},
		&stubProfileRepo{profiles: map[string]*domain.Profile{"prof-1": profile}},
	)
	// 23:00 UTC — should match the sleep activity
	now := time.Date(2024, 1, 15, 23, 0, 0, 0, time.UTC)
	state, err := svc.Resolve(context.Background(), "tok2", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.ProfileID != "prof-1" {
		t.Errorf("expected profile prof-1, got %q", state.ProfileID)
	}
	if state.ActiveActivity == nil || state.ActiveActivity.ID != "act-1" {
		t.Error("expected sleep activity to be active at 23:00")
	}
}

func TestClockService_scheduleTakesPrecedenceOverDefault(t *testing.T) {
	child := &domain.Child{
		ID:               "child-3",
		ClockToken:       "tok3",
		Timezone:         "UTC",
		DefaultProfileID: "prof-default",
	}
	defaultProfile := &domain.Profile{ID: "prof-default", Name: "Default"}
	schedProfile := &domain.Profile{
		ID:   "prof-sched",
		Name: "Monday",
		Activities: []domain.Activity{
			{ID: "act-school", Label: "School", FromHour: 8, ToHour: 15, Ring: 1},
		},
	}
	// Monday = weekday 1
	monday := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC) // 2024-01-15 is Monday
	assignment := &domain.DayAssignment{ChildID: "child-3", DayOfWeek: 1, ProfileID: "prof-sched"}
	svc := newSvc(
		&stubChildRepo{byToken: map[string]*domain.Child{"tok3": child}},
		&stubEventRepo{},
		&stubScheduleRepo{days: map[string]*domain.DayAssignment{
			scheduleKey("child-3", 1): assignment,
		}},
		&stubProfileRepo{profiles: map[string]*domain.Profile{
			"prof-default": defaultProfile,
			"prof-sched":   schedProfile,
		}},
	)
	state, err := svc.Resolve(context.Background(), "tok3", monday)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.ProfileID != "prof-sched" {
		t.Errorf("expected schedule profile, got %q", state.ProfileID)
	}
	if state.ActiveActivity == nil || state.ActiveActivity.ID != "act-school" {
		t.Error("expected school activity to be active at 10:00")
	}
}

func TestClockService_eventTakesPrecedenceOverSchedule(t *testing.T) {
	child := &domain.Child{
		ID:               "child-4",
		ClockToken:       "tok4",
		Timezone:         "UTC",
		DefaultProfileID: "prof-default",
	}
	schedProfile := &domain.Profile{ID: "prof-sched", Name: "Monday"}
	eventDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	event := domain.Event{
		ID:       "evt-1",
		ChildID:  "child-4",
		Date:     eventDate,
		FromTime: "09:00:00",
		ToTime:   "11:00:00",
		Activities: []domain.EventActivity{
			{ID: "ea-1", Label: "Doctor", FromHour: 9, ToHour: 11, Ring: 1},
		},
	}
	assignment := &domain.DayAssignment{ChildID: "child-4", DayOfWeek: 1, ProfileID: "prof-sched"}
	svc := newSvc(
		&stubChildRepo{byToken: map[string]*domain.Child{"tok4": child}},
		&stubEventRepo{events: []domain.Event{event}},
		&stubScheduleRepo{days: map[string]*domain.DayAssignment{
			scheduleKey("child-4", 1): assignment,
		}},
		&stubProfileRepo{profiles: map[string]*domain.Profile{"prof-sched": schedProfile}},
	)
	// 10:00 UTC on 2024-01-15 — event window 09:00–11:00 is active
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	state, err := svc.Resolve(context.Background(), "tok4", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.ActiveActivity == nil || state.ActiveActivity.ID != "ea-1" {
		t.Error("expected event activity to be active")
	}
}

func TestClockService_timezone(t *testing.T) {
	child := &domain.Child{
		ID:               "child-5",
		ClockToken:       "tok5",
		Timezone:         "America/New_York", // UTC-5 in winter
		DefaultProfileID: "prof-ny",
	}
	profile := &domain.Profile{
		ID:   "prof-ny",
		Name: "NY Profile",
		Activities: []domain.Activity{
			{ID: "act-morning", Label: "Breakfast", FromHour: 7, ToHour: 9, Ring: 1},
		},
	}
	svc := newSvc(
		&stubChildRepo{byToken: map[string]*domain.Child{"tok5": child}},
		&stubEventRepo{},
		&stubScheduleRepo{days: map[string]*domain.DayAssignment{}},
		&stubProfileRepo{profiles: map[string]*domain.Profile{"prof-ny": profile}},
	)
	// 12:00 UTC = 07:00 New York (UTC-5 in January)
	now := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	state, err := svc.Resolve(context.Background(), "tok5", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.ActiveActivity == nil || state.ActiveActivity.ID != "act-morning" {
		t.Errorf("expected morning activity at 07:00 NY time, got %+v", state.ActiveActivity)
	}
}
