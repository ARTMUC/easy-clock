package token_test

import (
	"testing"
	"time"

	"easy-clock/internal/token"
)

var secret = []byte("test-secret-key")

func TestSign_Validate_roundtrip(t *testing.T) {
	signed, err := token.Sign("user-123", secret, token.AccessTTL)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	userID, err := token.Validate(signed, secret)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if userID != "user-123" {
		t.Errorf("got userID %q, want %q", userID, "user-123")
	}
}

func TestValidate_expiredToken(t *testing.T) {
	signed, err := token.Sign("user-456", secret, -time.Second)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	_, err = token.Validate(signed, secret)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidate_wrongSecret(t *testing.T) {
	signed, err := token.Sign("user-789", secret, token.AccessTTL)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	_, err = token.Validate(signed, []byte("wrong-secret"))
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidate_malformedToken(t *testing.T) {
	_, err := token.Validate("not.a.jwt", secret)
	if err == nil {
		t.Fatal("expected error for malformed token, got nil")
	}
}
