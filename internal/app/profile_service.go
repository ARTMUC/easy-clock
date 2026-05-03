package app

import (
	"context"
	"fmt"

	"easy-clock/internal/domain"
)

type ProfileService struct {
	profileRepo  domain.ProfileRepository
	activityRepo domain.ActivityRepository
	presetRepo   domain.PresetActivityRepository
	childRepo    domain.ChildRepository
}

func NewProfileService(
	profileRepo domain.ProfileRepository,
	activityRepo domain.ActivityRepository,
	presetRepo domain.PresetActivityRepository,
	childRepo domain.ChildRepository,
) *ProfileService {
	return &ProfileService{
		profileRepo:  profileRepo,
		activityRepo: activityRepo,
		presetRepo:   presetRepo,
		childRepo:    childRepo,
	}
}

func (s *ProfileService) CreateProfile(ctx context.Context, childID, userID, name, color string) (*domain.Profile, error) {
	if err := s.assertChildOwner(ctx, childID, userID); err != nil {
		return nil, fmt.Errorf("ProfileService.CreateProfile: %w", err)
	}
	p, err := domain.NewProfile(childID, name, color)
	if err != nil {
		return nil, fmt.Errorf("ProfileService.CreateProfile: %w", err)
	}
	if err := s.profileRepo.Save(ctx, p); err != nil {
		return nil, fmt.Errorf("ProfileService.CreateProfile: save: %w", err)
	}
	return p, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, profileID, userID, name, color string) (*domain.Profile, error) {
	p, err := s.profileRepo.FindByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("ProfileService.UpdateProfile: find: %w", err)
	}
	if err := s.assertChildOwner(ctx, p.ChildID, userID); err != nil {
		return nil, fmt.Errorf("ProfileService.UpdateProfile: %w", err)
	}
	p.Name = name
	p.Color = color
	if err := s.profileRepo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("ProfileService.UpdateProfile: update: %w", err)
	}
	return p, nil
}

func (s *ProfileService) DeleteProfile(ctx context.Context, profileID, userID string) error {
	p, err := s.profileRepo.FindByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("ProfileService.DeleteProfile: find: %w", err)
	}
	if err := s.assertChildOwner(ctx, p.ChildID, userID); err != nil {
		return fmt.Errorf("ProfileService.DeleteProfile: %w", err)
	}
	if err := s.profileRepo.Delete(ctx, profileID); err != nil {
		return fmt.Errorf("ProfileService.DeleteProfile: delete: %w", err)
	}
	return nil
}

func (s *ProfileService) ListProfiles(ctx context.Context, childID, userID string) ([]domain.Profile, error) {
	if err := s.assertChildOwner(ctx, childID, userID); err != nil {
		return nil, fmt.Errorf("ProfileService.ListProfiles: %w", err)
	}
	profiles, err := s.profileRepo.FindByChildID(ctx, childID)
	if err != nil {
		return nil, fmt.Errorf("ProfileService.ListProfiles: %w", err)
	}
	return profiles, nil
}

func (s *ProfileService) GetProfile(ctx context.Context, profileID, userID string) (*domain.Profile, error) {
	p, err := s.profileRepo.FindWithActivities(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("ProfileService.GetProfile: find: %w", err)
	}
	if err := s.assertChildOwner(ctx, p.ChildID, userID); err != nil {
		return nil, fmt.Errorf("ProfileService.GetProfile: %w", err)
	}
	return p, nil
}

type AddActivityInput struct {
	PresetID  string // empty for custom
	Emoji     string
	Label     string
	ImagePath string
	FromHour  int
	ToHour    int
	Ring      int
	SortOrder int
}

func (s *ProfileService) AddActivity(ctx context.Context, profileID, userID string, in AddActivityInput) (*domain.Activity, error) {
	p, err := s.profileRepo.FindWithActivities(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("ProfileService.AddActivity: find profile: %w", err)
	}
	if err := s.assertChildOwner(ctx, p.ChildID, userID); err != nil {
		return nil, fmt.Errorf("ProfileService.AddActivity: %w", err)
	}
	if in.PresetID != "" && in.ImagePath == "" {
		preset, err := s.presetRepo.FindByID(ctx, in.PresetID)
		if err != nil {
			return nil, fmt.Errorf("ProfileService.AddActivity: find preset: %w", err)
		}
		in.ImagePath = preset.ImagePath
	}
	a, err := domain.NewActivity(profileID, in.PresetID, in.Emoji, in.Label, in.ImagePath, in.FromHour, in.ToHour, in.Ring, in.SortOrder)
	if err != nil {
		return nil, fmt.Errorf("ProfileService.AddActivity: %w", err)
	}
	if err := p.AddActivity(*a); err != nil {
		return nil, fmt.Errorf("ProfileService.AddActivity: overlap check: %w", err)
	}
	if err := s.activityRepo.Save(ctx, a); err != nil {
		return nil, fmt.Errorf("ProfileService.AddActivity: save: %w", err)
	}
	return a, nil
}

func (s *ProfileService) UpdateActivity(ctx context.Context, activityID, userID string, in AddActivityInput) (*domain.Activity, error) {
	a, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return nil, fmt.Errorf("ProfileService.UpdateActivity: find activity: %w", err)
	}
	p, err := s.profileRepo.FindWithActivities(ctx, a.ProfileID)
	if err != nil {
		return nil, fmt.Errorf("ProfileService.UpdateActivity: find profile: %w", err)
	}
	if err := s.assertChildOwner(ctx, p.ChildID, userID); err != nil {
		return nil, fmt.Errorf("ProfileService.UpdateActivity: %w", err)
	}
	tmp := domain.Profile{ChildID: p.ChildID}
	for _, existing := range p.Activities {
		if existing.ID == activityID {
			continue
		}
		_ = tmp.AddActivity(existing)
	}
	candidate := *a
	candidate.Emoji = in.Emoji
	candidate.Label = in.Label
	candidate.ImagePath = in.ImagePath
	candidate.FromHour = in.FromHour
	candidate.ToHour = in.ToHour
	candidate.Ring = in.Ring
	candidate.SortOrder = in.SortOrder
	if err := tmp.AddActivity(candidate); err != nil {
		return nil, fmt.Errorf("ProfileService.UpdateActivity: overlap check: %w", err)
	}
	*a = candidate
	if err := s.activityRepo.Update(ctx, a); err != nil {
		return nil, fmt.Errorf("ProfileService.UpdateActivity: update: %w", err)
	}
	return a, nil
}

func (s *ProfileService) RemoveActivity(ctx context.Context, activityID, userID string) error {
	a, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return fmt.Errorf("ProfileService.RemoveActivity: find activity: %w", err)
	}
	p, err := s.profileRepo.FindByID(ctx, a.ProfileID)
	if err != nil {
		return fmt.Errorf("ProfileService.RemoveActivity: find profile: %w", err)
	}
	if err := s.assertChildOwner(ctx, p.ChildID, userID); err != nil {
		return fmt.Errorf("ProfileService.RemoveActivity: %w", err)
	}
	if err := s.activityRepo.Delete(ctx, activityID); err != nil {
		return fmt.Errorf("ProfileService.RemoveActivity: delete: %w", err)
	}
	return nil
}

func (s *ProfileService) assertChildOwner(ctx context.Context, childID, userID string) error {
	c, err := s.childRepo.FindByID(ctx, childID)
	if err != nil {
		return fmt.Errorf("assertChildOwner: find child: %w", err)
	}
	if c.UserID != userID {
		return fmt.Errorf("assertChildOwner: %w", domain.ErrNotFound)
	}
	return nil
}
