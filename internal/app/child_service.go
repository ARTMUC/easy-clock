package app

import (
	"context"
	"errors"

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
		return nil, err
	}
	if err := s.childRepo.Save(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ChildService) UpdateChild(ctx context.Context, childID, userID, name, timezone, avatarPath string) (*domain.Child, error) {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return nil, err
	}
	if c.UserID != userID {
		return nil, domain.ErrNotFound
	}
	if _, err := domain.NewChild(userID, name, timezone); err != nil {
		return nil, err
	}
	c.Name = name
	c.Timezone = timezone
	c.AvatarPath = avatarPath
	if err := s.childRepo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ChildService) RemoveChild(ctx context.Context, childID, userID string) error {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return domain.ErrNotFound
	}
	return s.childRepo.Delete(ctx, childID)
}

func (s *ChildService) SetDefaultProfile(ctx context.Context, childID, userID, profileID string) error {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return err
	}
	if c.UserID != userID {
		return domain.ErrNotFound
	}
	profile, err := s.profileRepo.FindByID(ctx, profileID)
	if err != nil {
		return err
	}
	if profile.ChildID != childID {
		return errors.New("profile does not belong to this child")
	}
	c.DefaultProfileID = profileID
	return s.childRepo.Update(ctx, c)
}

func (s *ChildService) ListChildren(ctx context.Context, userID string) ([]domain.Child, error) {
	return s.childRepo.FindByUserID(ctx, userID)
}

func (s *ChildService) GetChild(ctx context.Context, childID, userID string) (*domain.Child, error) {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return nil, err
	}
	if c.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return c, nil
}
