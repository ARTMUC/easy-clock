package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"starter/internal/domain"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 30 * 24 * time.Hour
)

type UserService struct {
	userRepo    domain.UserRepository
	tokenRepo   domain.RefreshTokenRepository
}

func NewUserService(userRepo domain.UserRepository, tokenRepo domain.RefreshTokenRepository) *UserService {
	return &UserService{userRepo: userRepo, tokenRepo: tokenRepo}
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
}

func (s *UserService) Register(ctx context.Context, email, password string) (*domain.User, error) {
	if _, err := s.userRepo.FindByEmail(ctx, email); !errors.Is(err, domain.ErrNotFound) {
		if err == nil {
			return nil, domain.ErrEmailTaken
		}
		return nil, err
	}
	u, err := domain.NewUser(email, password)
	if err != nil {
		return nil, err
	}
	if err := s.userRepo.Save(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	u, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if !u.CheckPassword(password) {
		return nil, domain.ErrInvalidCredentials
	}
	return s.issueTokens(ctx, u.ID)
}

func (s *UserService) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := hashToken(rawRefreshToken)
	return s.tokenRepo.DeleteByHash(ctx, hash)
}

func (s *UserService) Refresh(ctx context.Context, rawRefreshToken string) (*LoginResult, error) {
	hash := hashToken(rawRefreshToken)
	stored, err := s.tokenRepo.FindByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if time.Now().UTC().After(stored.ExpiresAt) {
		_ = s.tokenRepo.DeleteByHash(ctx, hash)
		return nil, domain.ErrInvalidCredentials
	}
	_ = s.tokenRepo.DeleteByHash(ctx, hash)
	return s.issueTokens(ctx, stored.UserID)
}

func (s *UserService) issueTokens(ctx context.Context, userID string) (*LoginResult, error) {
	accessToken, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, err
	}
	rawRefresh, err := s.generateRefreshToken(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &LoginResult{AccessToken: accessToken, RefreshToken: rawRefresh}, nil
}

func (s *UserService) generateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().UTC().Add(accessTokenTTL).Unix(),
		"iat": time.Now().UTC().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret()))
}

func (s *UserService) generateRefreshToken(ctx context.Context, userID string) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	plain := hex.EncodeToString(raw)
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	rt := &domain.RefreshToken{
		ID:        id.String(),
		UserID:    userID,
		TokenHash: hashToken(plain),
		ExpiresAt: now.Add(refreshTokenTTL),
		CreatedAt: now,
	}
	if err := s.tokenRepo.Save(ctx, rt); err != nil {
		return "", err
	}
	return plain, nil
}

func hashToken(plain string) string {
	b, _ := hex.DecodeString(plain)
	if len(b) == 0 {
		b = []byte(plain)
	}
	sum := make([]byte, 32)
	copy(sum, b)
	return hex.EncodeToString(sum)
}

func jwtSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "change-me-in-production"
}
