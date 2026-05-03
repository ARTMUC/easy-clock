package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const AccessTTL = 15 * time.Minute

type claims struct {
	jwt.RegisteredClaims
}

func Sign(userID string, secret []byte, ttl time.Duration) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	s, err := t.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("sign: %w", err)
	}
	return s, nil
}

func Validate(tokenStr string, secret []byte) (userID string, err error) {
	t, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return "", err
	}
	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid {
		return "", fmt.Errorf("invalid token claims")
	}
	return c.Subject, nil
}
