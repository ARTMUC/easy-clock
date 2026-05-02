package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Version      int
	CreatedAt    time.Time // UTC
	UpdatedAt    time.Time // UTC
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time // UTC
	CreatedAt time.Time // UTC
}

func NewUser(email, plainPassword string) (*User, error) {
	if email == "" {
		return nil, ErrEmptyEmail
	}
	if len(plainPassword) < 8 {
		return nil, ErrPasswordTooShort
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	return &User{
		ID:           id.String(),
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) CheckPassword(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(plain)) == nil
}
