package user

import "context"

// Repository is the port (interface) for User persistence.
// The infrastructure layer provides the concrete adapter.
type Repository interface {
	Save(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByVerificationToken(ctx context.Context, token string) (*User, error)
}

// RefreshTokenRepository is the port for RefreshToken persistence.
type RefreshTokenRepository interface {
	Save(ctx context.Context, rt *RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Delete(ctx context.Context, id string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}
