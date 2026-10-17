package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

func signingKey() ([]byte, error) {
	value := os.Getenv("JWT_SECRET")
	if len(value) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must have at least 32 characters")
	}
	if os.Getenv("APP_ENV") == "production" && (strings.Contains(strings.ToLower(value), "demo") || strings.Contains(strings.ToLower(value), "change") || strings.Contains(value, "your-secret")) {
		return nil, fmt.Errorf("dummy signing keys are forbidden in production")
	}
	return []byte(value), nil
}
func ValidateSigningConfiguration() error { _, err := signingKey(); return err }
func ValidateJWT(value string) (*Claims, error) {
	key, err := signingKey()
	if err != nil {
		return nil, err
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(value, claims, func(_ *jwt.Token) (interface{}, error) { return key, nil }, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	if !token.Valid || claims.UserID <= 0 || claims.ExpiresAt == nil {
		return nil, fmt.Errorf("invalid customer token")
	}
	return claims, nil
}

func GenerateJWT(userID int) (string, error) {
	key, err := signingKey()
	if err != nil {
		return "", err
	}
	if userID <= 0 {
		return "", fmt.Errorf("invalid user identity")
	}
	id := make([]byte, 16)
	if _, err = rand.Read(id); err != nil {
		return "", err
	}
	claims := Claims{UserID: userID, RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), IssuedAt: jwt.NewNumericDate(time.Now()), ID: hex.EncodeToString(id)}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}
