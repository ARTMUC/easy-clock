package app

import (
	"context"
	"fmt"
	"time"

	"easy-clock/internal/domain"
)

type EventService struct {
	eventRepo   domain.EventRepository
	profileRepo domain.ProfileRepository
	childRepo   domain.ChildRepository
}

func NewEventService(
	eventRepo domain.EventRepository,
	profileRepo domain.ProfileRepository,
	childRepo domain.ChildRepository,
) *EventService {
	return &EventService{
		eventRepo:   eventRepo,
		profileRepo: profileRepo,
		childRepo:   childRepo,
	}
}

type CreateEventInput struct {
	Date       time.Time
	FromTime   string // "HH:MM:SS", LOCAL in child's timezone
	ToTime     string // "HH:MM:SS", LOCAL in child's timezone
	Label      string
	Emoji      string
	ProfileID  string // XOR with Activities
	Activities []EventActivityInput
}

type EventActivityInput struct {
	Emoji     string
	Label     string
	FromHour  int
	ToHour    int
	ImagePath string
}

func (s *EventService) CreateEvent(ctx context.Context, childID, userID string, in CreateEventInput) (*domain.Event, error) {
	if err := s.assertChildOwner(ctx, childID, userID); err != nil {
		return nil, fmt.Errorf("EventService.CreateEvent: %w", err)
	}
	e, err := domain.NewEvent(childID, in.Date, in.FromTime, in.ToTime, in.Label)
	if err != nil {
		return nil, fmt.Errorf("EventService.CreateEvent: %w", err)
	}
	e.Emoji = in.Emoji
	if in.ProfileID != "" {
		if err := e.SetProfile(in.ProfileID); err != nil {
			return nil, fmt.Errorf("EventService.CreateEvent: set profile: %w", err)
		}
		profile, err := s.profileRepo.FindByID(ctx, in.ProfileID)
		if err != nil {
			return nil, fmt.Errorf("EventService.CreateEvent: find profile: %w", err)
		}
		if profile.ChildID != childID {
			return nil, fmt.Errorf("EventService.CreateEvent: profile does not belong to child: %w", domain.ErrNotFound)
		}
	}
	for _, ai := range in.Activities {
		ea, err := s.buildEventActivity(e.ID, ai)
		if err != nil {
			return nil, fmt.Errorf("EventService.CreateEvent: build activity: %w", err)
		}
		if err := e.AddActivity(*ea); err != nil {
			return nil, fmt.Errorf("EventService.CreateEvent: add activity: %w", err)
		}
	}
	if err := s.eventRepo.Save(ctx, e); err != nil {
		return nil, fmt.Errorf("EventService.CreateEvent: save: %w", err)
	}
	return e, nil
}

func (s *EventService) UpdateEvent(ctx context.Context, eventID, userID string, in CreateEventInput) (*domain.Event, error) {
	e, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("EventService.UpdateEvent: find: %w", err)
	}
	if err := s.assertChildOwner(ctx, e.ChildID, userID); err != nil {
		return nil, fmt.Errorf("EventService.UpdateEvent: %w", err)
	}
	if in.FromTime >= in.ToTime {
		return nil, fmt.Errorf("EventService.UpdateEvent: %w", domain.ErrInvalidTimeRange)
	}
	e.Date = in.Date
	e.FromTime = in.FromTime
	e.ToTime = in.ToTime
	e.Label = in.Label
	e.Emoji = in.Emoji
	e.ProfileID = ""
	e.Activities = nil
	if in.ProfileID != "" {
		if err := e.SetProfile(in.ProfileID); err != nil {
			return nil, fmt.Errorf("EventService.UpdateEvent: set profile: %w", err)
		}
	}
	for _, ai := range in.Activities {
		ea, err := s.buildEventActivity(e.ID, ai)
		if err != nil {
			return nil, fmt.Errorf("EventService.UpdateEvent: build activity: %w", err)
		}
		if err := e.AddActivity(*ea); err != nil {
			return nil, fmt.Errorf("EventService.UpdateEvent: add activity: %w", err)
		}
	}
	if err := s.eventRepo.Update(ctx, e); err != nil {
		return nil, fmt.Errorf("EventService.UpdateEvent: update: %w", err)
	}
	return e, nil
}

func (s *EventService) DeleteEvent(ctx context.Context, eventID, userID string) error {
	e, err := s.eventRepo.FindByID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("EventService.DeleteEvent: find: %w", err)
	}
	if err := s.assertChildOwner(ctx, e.ChildID, userID); err != nil {
		return fmt.Errorf("EventService.DeleteEvent: %w", err)
	}
	if err := s.eventRepo.Delete(ctx, eventID); err != nil {
		return fmt.Errorf("EventService.DeleteEvent: delete: %w", err)
	}
	return nil
}

func (s *EventService) ListEvents(ctx context.Context, childID, userID string, from, to time.Time) ([]domain.Event, error) {
	if err := s.assertChildOwner(ctx, childID, userID); err != nil {
		return nil, fmt.Errorf("EventService.ListEvents: %w", err)
	}
	events, err := s.eventRepo.FindByChildID(ctx, childID, from, to)
	if err != nil {
		return nil, fmt.Errorf("EventService.ListEvents: %w", err)
	}
	return events, nil
}

func (s *EventService) buildEventActivity(eventID string, in EventActivityInput) (*domain.EventActivity, error) {
	id, err := newUUID()
	if err != nil {
		return nil, fmt.Errorf("buildEventActivity: uuid: %w", err)
	}
	now := time.Now()
	return &domain.EventActivity{
		ID:        id,
		EventID:   eventID,
		Emoji:     in.Emoji,
		Label:     in.Label,
		FromHour:  in.FromHour,
		ToHour:    in.ToHour,
		ImagePath: in.ImagePath,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (s *EventService) assertChildOwner(ctx context.Context, childID, userID string) error {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return fmt.Errorf("assertChildOwner: find child: %w", err)
	}
	if c.UserID != userID {
		return fmt.Errorf("assertChildOwner: %w", domain.ErrNotFound)
	}
	return nil
}
