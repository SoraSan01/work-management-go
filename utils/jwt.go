package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateAccessToken(userID, role, name, email string, permissions []string) (string, error) {
	claims := jwt.MapClaims{
		"sub":         userID,
		"role":        role,
		"name":        name,
		"email":       email,
		"permissions": permissions,
		"exp":         time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func GenerateRefreshToken(userID string) (string, time.Time) {
	exp := time.Now().Add(7 * 24 * time.Hour)
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return t, exp
}

func GenerateResetToken(userID string) (string, time.Time) {
	exp := time.Now().Add(1 * time.Hour) // 1 hour expiry
	token := uuid.NewString()            // or JWT
	return token, exp
}
