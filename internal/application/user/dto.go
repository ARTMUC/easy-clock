package userapplication

import domainuser "starter/internal/domain/user"

// RegisterRequest carries data needed to register a user with email/password.
type RegisterRequest struct {
	Name     string `form:"name"     binding:"required"`
	Email    string `form:"email"    binding:"required"`
	Password string `form:"password" binding:"required,min=8"`
}

// LoginRequest carries credentials for authentication.
type LoginRequest struct {
	Email    string `form:"email"    binding:"required"`
	Password string `form:"password" binding:"required"`
}

// UserDTO is the read representation of a User entity.
type UserDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func ToUserDTO(u *domainuser.User) UserDTO {
	return UserDTO{
		ID:    u.ID(),
		Name:  u.Name(),
		Email: u.Email(),
	}
}
