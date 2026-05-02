package app

import (
	"context"

	"starter/internal/domain"
)

type ScheduleService struct {
	scheduleRepo domain.ScheduleRepository
	profileRepo  domain.ProfileRepository
	childRepo    domain.ChildRepository
}

func NewScheduleService(
	scheduleRepo domain.ScheduleRepository,
	profileRepo domain.ProfileRepository,
	childRepo domain.ChildRepository,
) *ScheduleService {
	return &ScheduleService{
		scheduleRepo: scheduleRepo,
		profileRepo:  profileRepo,
		childRepo:    childRepo,
	}
}

func (s *ScheduleService) GetSchedule(ctx context.Context, childID, userID string) ([]domain.DayAssignment, error) {
	if err := s.assertChildOwner(ctx, childID, userID); err != nil {
		return nil, err
	}
	return s.scheduleRepo.FindByChildID(ctx, childID)
}

// AssignProfileToDays bulk-assigns one profile to multiple days of the week.
func (s *ScheduleService) AssignProfileToDays(ctx context.Context, childID, userID, profileID string, days []int) error {
	if err := s.assertChildOwner(ctx, childID, userID); err != nil {
		return err
	}
	profile, err := s.profileRepo.FindByID(ctx, profileID)
	if err != nil {
		return err
	}
	if profile.ChildID != childID {
		return domain.ErrNotFound
	}
	assignments := make([]domain.DayAssignment, 0, len(days))
	for _, day := range days {
		a, err := domain.NewDayAssignment(childID, day, profileID)
		if err != nil {
			return err
		}
		assignments = append(assignments, *a)
	}
	return s.scheduleRepo.UpsertMany(ctx, assignments)
}

func (s *ScheduleService) ClearDay(ctx context.Context, childID, userID string, day int) error {
	if err := s.assertChildOwner(ctx, childID, userID); err != nil {
		return err
	}
	return s.scheduleRepo.DeleteDay(ctx, childID, day)
}

func (s *ScheduleService) assertChildOwner(ctx context.Context, childID, userID string) error {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return domain.ErrNotFound
	}
	return nil
}
