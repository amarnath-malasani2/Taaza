package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func signingKey() ([]byte, error) {
	value := os.Getenv("ADMIN_JWT_SECRET")
	if len(value) < 32 {
		return nil, fmt.Errorf("ADMIN_JWT_SECRET must have at least 32 characters")
	}
	if os.Getenv("APP_ENV") == "production" && (strings.Contains(strings.ToLower(value), "demo") || strings.Contains(strings.ToLower(value), "change")) {
		return nil, fmt.Errorf("dummy admin signing keys are forbidden in production")
	}
	return []byte(value), nil
}
func ValidateSigningConfiguration() error { _, err := signingKey(); return err }

type AdminClaims struct {
	AdminID int    `json:"admin_id"`
	Email   string `json:"email"`
	Role    string `json:"role"`
	jwt.RegisteredClaims
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateAdminJWT generates a JWT token for an admin
func GenerateAdminJWT(adminID int, email, role string) (string, error) {
	key, err := signingKey()
	if err != nil {
		return "", err
	}
	id := make([]byte, 16)
	if _, err = rand.Read(id); err != nil {
		return "", err
	}
	claims := AdminClaims{
		AdminID: adminID,
		Email:   email,
		Role:    role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        hex.EncodeToString(id),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(key)
}

// ValidateAdminJWT validates an admin JWT token
func ValidateAdminJWT(tokenString string) (*AdminClaims, error) {
	key, err := signingKey()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid && claims.AdminID > 0 && claims.ExpiresAt != nil {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
