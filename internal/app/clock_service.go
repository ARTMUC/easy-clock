package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"easy-clock/internal/domain"
)

type ClockService struct {
	childRepo    domain.ChildRepository
	eventRepo    domain.EventRepository
	scheduleRepo domain.ScheduleRepository
	profileRepo  domain.ProfileRepository
}

func NewClockService(
	childRepo domain.ChildRepository,
	eventRepo domain.EventRepository,
	scheduleRepo domain.ScheduleRepository,
	profileRepo domain.ProfileRepository,
) *ClockService {
	return &ClockService{
		childRepo:    childRepo,
		eventRepo:    eventRepo,
		scheduleRepo: scheduleRepo,
		profileRepo:  profileRepo,
	}
}

// Resolve returns the ClockState for the child identified by clockToken at the given UTC time.
// Priority: Event > DayAssignment > DefaultProfile.
func (s *ClockService) Resolve(ctx context.Context, clockToken string, now time.Time) (*domain.ClockState, error) {
	child, err := s.childRepo.FindByClockToken(ctx, clockToken)
	if err != nil {
		return nil, fmt.Errorf("ClockService.Resolve: find child: %w", err)
	}

	loc, err := time.LoadLocation(child.Timezone)
	if err != nil {
		return nil, fmt.Errorf("ClockService.Resolve: load timezone %q: %w", child.Timezone, err)
	}
	localNow := now.In(loc)

	// 1. One-off events — highest priority.
	events, err := s.eventRepo.FindForDate(ctx, child.ID, localNow.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("ClockService.Resolve: find events: %w", err)
	}
	for i := range events {
		e := &events[i]
		if localTimeInRange(localNow, e.FromTime, e.ToTime) {
			return resolveFromEvent(e, localNow), nil
		}
	}

	// 2. Weekly schedule.
	weekday := int(localNow.Weekday())
	assignment, err := s.scheduleRepo.FindDay(ctx, child.ID, weekday)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("ClockService.Resolve: find schedule day: %w", err)
	}
	if assignment != nil {
		profile, err := s.profileRepo.FindWithActivities(ctx, assignment.ProfileID)
		if err != nil {
			return nil, fmt.Errorf("ClockService.Resolve: find schedule profile: %w", err)
		}
		return resolveFromProfile(profile, localNow), nil
	}

	// 3. Default profile fallback.
	if child.DefaultProfileID != "" {
		profile, err := s.profileRepo.FindWithActivities(ctx, child.DefaultProfileID)
		if err != nil {
			return nil, fmt.Errorf("ClockService.Resolve: find default profile: %w", err)
		}
		return resolveFromProfile(profile, localNow), nil
	}

	return &domain.ClockState{Empty: true}, nil
}

func resolveFromProfile(p *domain.Profile, localNow time.Time) *domain.ClockState {
	state := &domain.ClockState{
		ProfileID:     p.ID,
		ProfileName:   p.Name,
		ProfileColor:  p.Color,
		AllActivities: p.Activities,
	}
	h := localNow.Hour()
	for i := range p.Activities {
		a := &p.Activities[i]
		if h >= a.FromHour && h < a.ToHour {
			state.ActiveActivity = a
			break
		}
	}
	return state
}

func resolveFromEvent(e *domain.Event, localNow time.Time) *domain.ClockState {
	if e.ProfileID != "" {
		return &domain.ClockState{ProfileID: e.ProfileID}
	}
	activities := make([]domain.Activity, len(e.Activities))
	h := localNow.Hour()
	state := &domain.ClockState{}
	for i, ea := range e.Activities {
		a := domain.Activity{
			ID:        ea.ID,
			Emoji:     ea.Emoji,
			Label:     ea.Label,
			FromHour:  ea.FromHour,
			ToHour:    ea.ToHour,
			Ring:      ea.Ring,
			ImagePath: ea.ImagePath,
		}
		activities[i] = a
		if h >= ea.FromHour && h < ea.ToHour {
			state.ActiveActivity = &activities[i]
		}
	}
	state.AllActivities = activities
	return state
}

// localTimeInRange checks if localNow falls within [fromTime, toTime) where times are "HH:MM:SS".
func localTimeInRange(localNow time.Time, fromTime, toTime string) bool {
	base := localNow.Format("2006-01-02")
	loc := localNow.Location()
	from, err1 := time.ParseInLocation("2006-01-02 15:04:05", base+" "+fromTime, loc)
	to, err2 := time.ParseInLocation("2006-01-02 15:04:05", base+" "+toTime, loc)
	if err1 != nil || err2 != nil {
		return false
	}
	return !localNow.Before(from) && localNow.Before(to)
}
