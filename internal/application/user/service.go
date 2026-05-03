package userapplication

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	domainuser "easy-clock/internal/domain/user"
	"easy-clock/internal/eventbus"
	"easy-clock/internal/token"

	"github.com/google/uuid"
)

// Service is the application service for the User bounded context.
type Service struct {
	repo        domainuser.Repository
	refreshRepo domainuser.RefreshTokenRepository
	bus         *eventbus.Bus
	emailSender EmailSender
	baseURL     string
	jwtSecret   []byte
}

// NewService creates a new application service.
func NewService(
	repo domainuser.Repository,
	refreshRepo domainuser.RefreshTokenRepository,
	bus *eventbus.Bus,
	emailSender EmailSender,
	baseURL string,
	jwtSecret []byte,
) *Service {
	return &Service{
		repo:        repo,
		refreshRepo: refreshRepo,
		bus:         bus,
		emailSender: emailSender,
		baseURL:     baseURL,
		jwtSecret:   jwtSecret,
	}
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

	verifyToken, err := generateRandom()
	if err != nil {
		return UserDTO{}, fmt.Errorf("register: generate token: %w", err)
	}

	u, err := domainuser.NewUserWithAuth(req.Name, req.Email, string(hash), verifyToken)
	if err != nil {
		return UserDTO{}, fmt.Errorf("register: %w", err)
	}

	verifyURL := s.baseURL + "/verify?token=" + verifyToken
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

// Login authenticates a user and returns a plain UserDTO (used by browser session flow).
func (s *Service) Login(ctx context.Context, req LoginRequest) (UserDTO, error) {
	u, err := s.validateCredentials(ctx, req)
	if err != nil {
		return UserDTO{}, err
	}
	return ToUserDTO(u), nil
}

// LoginWithTokens authenticates a user and issues a JWT access token + refresh token.
func (s *Service) LoginWithTokens(ctx context.Context, req LoginRequest) (TokenPairDTO, error) {
	u, err := s.validateCredentials(ctx, req)
	if err != nil {
		return TokenPairDTO{}, err
	}
	return s.issueTokenPair(ctx, u.ID())
}

// Refresh validates a refresh token, rotates it, and issues a new token pair.
func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (TokenPairDTO, error) {
	hash := hashToken(rawRefreshToken)
	rt, err := s.refreshRepo.FindByHash(ctx, hash)
	if err != nil {
		return TokenPairDTO{}, domainuser.ErrInvalidToken
	}
	if time.Now().After(rt.ExpiresAt) {
		_ = s.refreshRepo.Delete(ctx, rt.ID)
		return TokenPairDTO{}, domainuser.ErrInvalidToken
	}
	if err := s.refreshRepo.Delete(ctx, rt.ID); err != nil {
		return TokenPairDTO{}, fmt.Errorf("rotate refresh token: %w", err)
	}
	return s.issueTokenPair(ctx, rt.UserID)
}

// RevokeToken invalidates a refresh token (logout).
func (s *Service) RevokeToken(ctx context.Context, rawRefreshToken string) error {
	hash := hashToken(rawRefreshToken)
	rt, err := s.refreshRepo.FindByHash(ctx, hash)
	if err != nil {
		return nil // idempotent — already gone
	}
	return s.refreshRepo.Delete(ctx, rt.ID)
}

// VerifyEmail activates the user account associated with the given token.
func (s *Service) VerifyEmail(ctx context.Context, verifyToken string) error {
	if verifyToken == "" {
		return domainuser.ErrInvalidToken
	}

	u, err := s.repo.FindByVerificationToken(ctx, verifyToken)
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

func (s *Service) validateCredentials(ctx context.Context, req LoginRequest) (*domainuser.User, error) {
	u, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, domainuser.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash()), []byte(req.Password)); err != nil {
		return nil, domainuser.ErrInvalidCredentials
	}
	if !u.Active() {
		return nil, domainuser.ErrNotActive
	}
	return u, nil
}

func (s *Service) issueTokenPair(ctx context.Context, userID string) (TokenPairDTO, error) {
	accessToken, err := token.Sign(userID, s.jwtSecret, token.AccessTTL)
	if err != nil {
		return TokenPairDTO{}, fmt.Errorf("sign access token: %w", err)
	}

	rawRefresh, err := generateRandom()
	if err != nil {
		return TokenPairDTO{}, fmt.Errorf("generate refresh token: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return TokenPairDTO{}, fmt.Errorf("generate refresh token id: %w", err)
	}

	rt := &domainuser.RefreshToken{
		ID:        id.String(),
		UserID:    userID,
		TokenHash: hashToken(rawRefresh),
		ExpiresAt: time.Now().Add(domainuser.RefreshTTL),
	}
	if err := s.refreshRepo.Save(ctx, rt); err != nil {
		return TokenPairDTO{}, fmt.Errorf("persist refresh token: %w", err)
	}

	return TokenPairDTO{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresIn:    int64(token.AccessTTL.Seconds()),
	}, nil
}

func generateRandom() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
