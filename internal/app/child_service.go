package app

import (
	"context"
	"fmt"

	"starter/internal/domain"
)

type ChildService struct {
	childRepo   domain.ChildRepository
	profileRepo domain.ProfileRepository
}

func NewChildService(childRepo domain.ChildRepository, profileRepo domain.ProfileRepository) *ChildService {
	return &ChildService{childRepo: childRepo, profileRepo: profileRepo}
}

func (s *ChildService) AddChild(ctx context.Context, userID, name, timezone string) (*domain.Child, error) {
	c, err := domain.NewChild(userID, name, timezone)
	if err != nil {
		return nil, fmt.Errorf("ChildService.AddChild: %w", err)
	}
	if err := s.childRepo.Save(ctx, c); err != nil {
		return nil, fmt.Errorf("ChildService.AddChild: save: %w", err)
	}
	return c, nil
}

func (s *ChildService) UpdateChild(ctx context.Context, childID, userID, name, timezone, avatarPath string) (*domain.Child, error) {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return nil, fmt.Errorf("ChildService.UpdateChild: find: %w", err)
	}
	if c.UserID != userID {
		return nil, fmt.Errorf("ChildService.UpdateChild: %w", domain.ErrNotFound)
	}
	if _, err := domain.NewChild(userID, name, timezone); err != nil {
		return nil, fmt.Errorf("ChildService.UpdateChild: validate: %w", err)
	}
	c.Name = name
	c.Timezone = timezone
	c.AvatarPath = avatarPath
	if err := s.childRepo.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("ChildService.UpdateChild: update: %w", err)
	}
	return c, nil
}

func (s *ChildService) RemoveChild(ctx context.Context, childID, userID string) error {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return fmt.Errorf("ChildService.RemoveChild: find: %w", err)
	}
	if c.UserID != userID {
		return fmt.Errorf("ChildService.RemoveChild: %w", domain.ErrNotFound)
	}
	if err := s.childRepo.Delete(ctx, childID); err != nil {
		return fmt.Errorf("ChildService.RemoveChild: delete: %w", err)
	}
	return nil
}

func (s *ChildService) SetDefaultProfile(ctx context.Context, childID, userID, profileID string) error {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return fmt.Errorf("ChildService.SetDefaultProfile: find child: %w", err)
	}
	if c.UserID != userID {
		return fmt.Errorf("ChildService.SetDefaultProfile: %w", domain.ErrNotFound)
	}
	profile, err := s.profileRepo.FindByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("ChildService.SetDefaultProfile: find profile: %w", err)
	}
	if profile.ChildID != childID {
		return fmt.Errorf("ChildService.SetDefaultProfile: profile does not belong to child: %w", domain.ErrNotFound)
	}
	c.DefaultProfileID = profileID
	if err := s.childRepo.Update(ctx, c); err != nil {
		return fmt.Errorf("ChildService.SetDefaultProfile: update: %w", err)
	}
	return nil
}

func (s *ChildService) ListChildren(ctx context.Context, userID string) ([]domain.Child, error) {
	children, err := s.childRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ChildService.ListChildren: %w", err)
	}
	return children, nil
}

func (s *ChildService) GetChild(ctx context.Context, childID, userID string) (*domain.Child, error) {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return nil, fmt.Errorf("ChildService.GetChild: %w", err)
	}
	if c.UserID != userID {
		return nil, fmt.Errorf("ChildService.GetChild: %w", domain.ErrNotFound)
	}
	return c, nil
}
