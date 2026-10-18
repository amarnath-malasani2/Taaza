package utils

import (
	"fmt"
	"os"
	"strings"

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
