package app

import (
	"context"
	"fmt"

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
		return nil, fmt.Errorf("ScheduleService.GetSchedule: %w", err)
	}
	assignments, err := s.scheduleRepo.FindByChildID(ctx, childID)
	if err != nil {
		return nil, fmt.Errorf("ScheduleService.GetSchedule: %w", err)
	}
	return assignments, nil
}

func (s *ScheduleService) AssignProfileToDays(ctx context.Context, childID, userID, profileID string, days []int) error {
	if err := s.assertChildOwner(ctx, childID, userID); err != nil {
		return fmt.Errorf("ScheduleService.AssignProfileToDays: %w", err)
	}
	profile, err := s.profileRepo.FindByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("ScheduleService.AssignProfileToDays: find profile: %w", err)
	}
	if profile.ChildID != childID {
		return fmt.Errorf("ScheduleService.AssignProfileToDays: profile does not belong to child: %w", domain.ErrNotFound)
	}
	assignments := make([]domain.DayAssignment, 0, len(days))
	for _, day := range days {
		a, err := domain.NewDayAssignment(childID, day, profileID)
		if err != nil {
			return fmt.Errorf("ScheduleService.AssignProfileToDays: build assignment day=%d: %w", day, err)
		}
		assignments = append(assignments, *a)
	}
	if err := s.scheduleRepo.UpsertMany(ctx, assignments); err != nil {
		return fmt.Errorf("ScheduleService.AssignProfileToDays: upsert: %w", err)
	}
	return nil
}

func (s *ScheduleService) ClearDay(ctx context.Context, childID, userID string, day int) error {
	if err := s.assertChildOwner(ctx, childID, userID); err != nil {
		return fmt.Errorf("ScheduleService.ClearDay: %w", err)
	}
	if err := s.scheduleRepo.DeleteDay(ctx, childID, day); err != nil {
		return fmt.Errorf("ScheduleService.ClearDay: %w", err)
	}
	return nil
}

func (s *ScheduleService) assertChildOwner(ctx context.Context, childID, userID string) error {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return fmt.Errorf("assertChildOwner: find child: %w", err)
	}
	if c.UserID != userID {
		return fmt.Errorf("assertChildOwner: %w", domain.ErrNotFound)
	}
	return nil
}
