package userapplication_test

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	userapplication "easy-clock/internal/application/user"
	domainuser "easy-clock/internal/domain/user"
	"easy-clock/internal/eventbus"
)

// ---- in-memory stubs ----

type stubUserRepo struct {
	byEmail map[string]*domainuser.User
	byToken map[string]*domainuser.User
}

func (r *stubUserRepo) Save(_ context.Context, u *domainuser.User) error {
	if r.byEmail == nil {
		r.byEmail = map[string]*domainuser.User{}
	}
	r.byEmail[u.Email()] = u
	return nil
}

func (r *stubUserRepo) FindByEmail(_ context.Context, email string) (*domainuser.User, error) {
	if u, ok := r.byEmail[email]; ok {
		return u, nil
	}
	return nil, domainuser.ErrNotFound
}

func (r *stubUserRepo) FindByVerificationToken(_ context.Context, tok string) (*domainuser.User, error) {
	if u, ok := r.byToken[tok]; ok {
		return u, nil
	}
	return nil, domainuser.ErrNotFound
}

type stubRefreshRepo struct {
	byHash map[string]*domainuser.RefreshToken
}

func (r *stubRefreshRepo) Save(_ context.Context, rt *domainuser.RefreshToken) error {
	if r.byHash == nil {
		r.byHash = map[string]*domainuser.RefreshToken{}
	}
	r.byHash[rt.TokenHash] = rt
	return nil
}

func (r *stubRefreshRepo) FindByHash(_ context.Context, hash string) (*domainuser.RefreshToken, error) {
	if rt, ok := r.byHash[hash]; ok {
		return rt, nil
	}
	return nil, domainuser.ErrNotFound
}

func (r *stubRefreshRepo) Delete(_ context.Context, id string) error {
	for h, rt := range r.byHash {
		if rt.ID == id {
			delete(r.byHash, h)
			return nil
		}
	}
	return nil
}

func (r *stubRefreshRepo) DeleteAllForUser(_ context.Context, userID string) error {
	for h, rt := range r.byHash {
		if rt.UserID == userID {
			delete(r.byHash, h)
		}
	}
	return nil
}

type noopEmail struct{}

func (noopEmail) SendVerificationEmail(_ context.Context, _, _, _ string) error { return nil }

// ---- helpers ----

func newService() (*userapplication.Service, *stubUserRepo, *stubRefreshRepo) {
	userRepo := &stubUserRepo{byEmail: map[string]*domainuser.User{}}
	refreshRepo := &stubRefreshRepo{byHash: map[string]*domainuser.RefreshToken{}}
	bus := eventbus.New()
	svc := userapplication.NewService(userRepo, refreshRepo, bus, noopEmail{}, "http://localhost", []byte("test-secret"))
	return svc, userRepo, refreshRepo
}

func bcryptHash(password string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

func activeUser(t *testing.T, email, password string) *domainuser.User {
	t.Helper()
	hash, err := bcryptHash(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	return domainuser.Reconstitute("user-id-1", "Test User", email, hash, true, "")
}

// ---- tests ----

func TestLogin_validCredentials(t *testing.T) {
	svc, userRepo, _ := newService()
	u := activeUser(t, "alice@example.com", "secret123")
	userRepo.byEmail["alice@example.com"] = u

	dto, err := svc.Login(context.Background(), userapplication.LoginRequest{
		Email:    "alice@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if dto.Email != "alice@example.com" {
		t.Errorf("got email %q", dto.Email)
	}
}

func TestLogin_invalidPassword(t *testing.T) {
	svc, userRepo, _ := newService()
	u := activeUser(t, "alice@example.com", "secret123")
	userRepo.byEmail["alice@example.com"] = u

	_, err := svc.Login(context.Background(), userapplication.LoginRequest{
		Email:    "alice@example.com",
		Password: "wrong-password",
	})
	if !errors.Is(err, domainuser.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_unknownEmail(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Login(context.Background(), userapplication.LoginRequest{
		Email:    "nobody@example.com",
		Password: "secret123",
	})
	if !errors.Is(err, domainuser.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_inactiveAccount(t *testing.T) {
	svc, userRepo, _ := newService()
	hash, _ := bcryptHash("secret123")
	u := domainuser.Reconstitute("user-id-2", "Bob", "bob@example.com", hash, false, "verify-tok")
	userRepo.byEmail["bob@example.com"] = u

	_, err := svc.Login(context.Background(), userapplication.LoginRequest{
		Email:    "bob@example.com",
		Password: "secret123",
	})
	if !errors.Is(err, domainuser.ErrNotActive) {
		t.Fatalf("expected ErrNotActive, got %v", err)
	}
}

func TestLoginWithTokens_issuesJWT(t *testing.T) {
	svc, userRepo, refreshRepo := newService()
	u := activeUser(t, "carol@example.com", "secret123")
	userRepo.byEmail["carol@example.com"] = u

	pair, err := svc.LoginWithTokens(context.Background(), userapplication.LoginRequest{
		Email:    "carol@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("LoginWithTokens: %v", err)
	}
	if pair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if pair.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if len(refreshRepo.byHash) != 1 {
		t.Errorf("expected 1 refresh token stored, got %d", len(refreshRepo.byHash))
	}
}

func TestRefresh_rotatesToken(t *testing.T) {
	svc, userRepo, refreshRepo := newService()
	u := activeUser(t, "dave@example.com", "secret123")
	userRepo.byEmail["dave@example.com"] = u

	first, err := svc.LoginWithTokens(context.Background(), userapplication.LoginRequest{
		Email:    "dave@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("LoginWithTokens: %v", err)
	}
	if len(refreshRepo.byHash) != 1 {
		t.Fatalf("expected 1 refresh token before rotation, got %d", len(refreshRepo.byHash))
	}

	second, err := svc.Refresh(context.Background(), first.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if second.RefreshToken == first.RefreshToken {
		t.Error("expected new refresh token after rotation")
	}
	if second.AccessToken == "" {
		t.Error("expected non-empty new access token")
	}
	if len(refreshRepo.byHash) != 1 {
		t.Errorf("expected 1 refresh token after rotation, got %d", len(refreshRepo.byHash))
	}
}

func TestRefresh_invalidToken(t *testing.T) {
	svc, _, _ := newService()
	_, err := svc.Refresh(context.Background(), "not-a-valid-token")
	if !errors.Is(err, domainuser.ErrInvalidToken) {
		t.Fatalf("expected ErrInvalidToken, got %v", err)
	}
}

func TestRevokeToken_idempotent(t *testing.T) {
	svc, userRepo, refreshRepo := newService()
	u := activeUser(t, "eve@example.com", "secret123")
	userRepo.byEmail["eve@example.com"] = u

	pair, _ := svc.LoginWithTokens(context.Background(), userapplication.LoginRequest{
		Email:    "eve@example.com",
		Password: "secret123",
	})

	if err := svc.RevokeToken(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("first revoke: %v", err)
	}
	if len(refreshRepo.byHash) != 0 {
		t.Errorf("expected 0 refresh tokens after revoke, got %d", len(refreshRepo.byHash))
	}
	// second call must not error
	if err := svc.RevokeToken(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("second revoke (idempotent): %v", err)
	}
}
