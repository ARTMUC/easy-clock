package userpersistence

import domainuser "starter/internal/domain/user"

// Translator maps between the User domain entity and the GORM persistence models.
type Translator struct{}

// ToModel maps a domain User to a UserWriteModel (used for INSERT / UPDATE).
func (t Translator) ToModel(u *domainuser.User) *UserWriteModel {
	return &UserWriteModel{
		ID:                u.ID(),
		Name:              u.Name(),
		Email:             u.Email(),
		PasswordHash:      u.PasswordHash(),
		Active:            u.Active(),
		VerificationToken: u.VerificationToken(),
	}
}

// ToDomain maps a UserModel (read model) back to a domain User.
func (t Translator) ToDomain(m *UserModel) *domainuser.User {
	return domainuser.Reconstitute(m.ID, m.Name, m.Email, m.PasswordHash, m.Active, m.VerificationToken)
}
