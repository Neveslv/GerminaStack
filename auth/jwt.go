package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	CookieName     = "token"
	ContextUserID  = "user_id"
	ContextIsAdmin = "is_admin"
)

type Claims struct {
	IsAdmin bool `json:"is_admin"`
	jwt.RegisteredClaims
}

func ParseToken(tokenString, secret string) (*Claims, error) {
	if secret == "" {
		return nil, errors.New("JWT secret is required")
	}
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return claims, nil
}

func GenerateToken(userID string, isAdmin bool, secret string) (string, error) {
	issuedAt := time.Now().UTC()
	return GenerateTokenWithTimes(userID, isAdmin, secret, issuedAt, issuedAt.Add(24*time.Hour))
}

func GenerateTokenWithTimes(userID string, isAdmin bool, secret string, issuedAt, expiresAt time.Time) (string, error) {
	if userID == "" {
		return "", errors.New("JWT subject is required")
	}
	if secret == "" {
		return "", errors.New("JWT secret is required")
	}
	issuedAt = issuedAt.UTC()
	expiresAt = expiresAt.UTC()
	if !expiresAt.After(issuedAt) {
		return "", errors.New("JWT expiration must be after issuance")
	}

	claims := &Claims{
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
