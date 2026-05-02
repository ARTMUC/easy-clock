package userapplication

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	domainuser "starter/internal/domain/user"
	"starter/internal/eventbus"
)

// Service is the application service for the User bounded context.
type Service struct {
	repo        domainuser.Repository
	bus         *eventbus.Bus
	emailSender EmailSender
	baseURL     string
}

// NewService creates a new application service.
func NewService(repo domainuser.Repository, bus *eventbus.Bus, emailSender EmailSender, baseURL string) *Service {
	return &Service{repo: repo, bus: bus, emailSender: emailSender, baseURL: baseURL}
}

// -----------------------------------------------------------------
// Use cases
// -----------------------------------------------------------------

// Register creates a new inactive user, persists it, and sends a verification email.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (UserDTO, error) {
	_, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil {
		return UserDTO{}, domainuser.ErrEmailTaken
	}
	if !errors.Is(err, domainuser.ErrNotFound) {
		return UserDTO{}, fmt.Errorf("register: check email: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return UserDTO{}, fmt.Errorf("register: hash password: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return UserDTO{}, fmt.Errorf("register: generate token: %w", err)
	}

	u, err := domainuser.NewUserWithAuth(req.Name, req.Email, string(hash), token)
	if err != nil {
		return UserDTO{}, fmt.Errorf("register: %w", err)
	}

	verifyURL := s.baseURL + "/verify?token=" + token
	if err := s.emailSender.SendVerificationEmail(ctx, u.Email(), u.Name(), verifyURL); err != nil {
		fmt.Printf("[EMAIL ERROR] %v\n", err)
		return UserDTO{}, fmt.Errorf("register: send verification email: %w", err)
	}

	if err := s.repo.Save(ctx, u); err != nil {
		return UserDTO{}, fmt.Errorf("register: persist: %w", err)
	}

	eventbus.PublishAll(s.bus, u.PullEvents())
	return ToUserDTO(u), nil
}

// Login authenticates a user. Returns ErrNotActive if the account is not yet verified.
func (s *Service) Login(ctx context.Context, req LoginRequest) (UserDTO, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return UserDTO{}, domainuser.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash()), []byte(req.Password)); err != nil {
		return UserDTO{}, domainuser.ErrInvalidCredentials
	}

	if !u.Active() {
		return UserDTO{}, domainuser.ErrNotActive
	}

	return ToUserDTO(u), nil
}

// VerifyEmail activates the user account associated with the given token.
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	if token == "" {
		return domainuser.ErrInvalidToken
	}

	u, err := s.repo.FindByVerificationToken(ctx, token)
	if err != nil {
		return domainuser.ErrInvalidToken
	}

	u.Activate()

	if err := s.repo.Save(ctx, u); err != nil {
		return fmt.Errorf("verify email: persist: %w", err)
	}
	return nil
}

// -----------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
