package user

import "errors"

var (
	ErrEmptyName          = errors.New("user name cannot be empty")
	ErrNotFound           = errors.New("user not found")
	ErrEmptyEmail         = errors.New("email cannot be empty")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrNotActive          = errors.New("account not activated — check your email")
	ErrInvalidToken       = errors.New("invalid or expired verification token")
)
