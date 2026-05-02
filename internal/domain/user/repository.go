package user

import "context"

// Repository is the port (interface) for User persistence.
// The infrastructure layer provides the concrete adapter.
type Repository interface {
	// Save persists a User. For a new User (ID == 0) it inserts and sets the ID.
	// For an existing User it updates the record.
	Save(ctx context.Context, u *User) error

	// FindByEmail loads a User by their email address.
	FindByEmail(ctx context.Context, email string) (*User, error)

	// FindByVerificationToken loads a User by their verification token.
	FindByVerificationToken(ctx context.Context, token string) (*User, error)
}
