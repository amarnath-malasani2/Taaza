package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestJWTRejectsInvalidClaimsAndAlgorithms(t *testing.T) {
	key := "test-signing-key-longer-than-thirty-two-characters"
	t.Setenv("JWT_SECRET", key)
	t.Setenv("APP_ENV", "test")
	for _, tc := range []struct {
		name    string
		method  jwt.SigningMethod
		expires *jwt.NumericDate
		valid   bool
	}{
		{"valid", jwt.SigningMethodHS256, jwt.NewNumericDate(time.Now().Add(time.Hour)), true},
		{"expired", jwt.SigningMethodHS256, jwt.NewNumericDate(time.Now().Add(-time.Hour)), false},
		{"missing expiry", jwt.SigningMethodHS256, nil, false},
		{"different algorithm", jwt.SigningMethodHS512, jwt.NewNumericDate(time.Now().Add(time.Hour)), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registered := jwt.RegisteredClaims{ExpiresAt: tc.expires}
			token, err := jwt.NewWithClaims(tc.method, Claims{UserID: 1, RegisteredClaims: registered}).SignedString([]byte(key))
			if err != nil {
				t.Fatal(err)
			}
			_, err = ValidateJWT(token)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
		})
	}
}
func TestProductionRejectsDummySigningKey(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("JWT_SECRET", "demo-only-signing-key-with-at-least-32-characters")
	if ValidateSigningConfiguration() == nil {
		t.Fatal("production accepted dummy key")
	}
	t.Setenv("JWT_SECRET", "short")
	if ValidateSigningConfiguration() == nil {
		t.Fatal("short key accepted")
	}
	t.Setenv("JWT_SECRET", "a7c3f9b1e8d2a5c6f0e4b9a8d3c7e2f1a6b5c9d8e0f2a4b7c1d3e5f6a8b0c2d4")
	if err := ValidateSigningConfiguration(); err != nil {
		t.Fatal(err)
	}
}
func TestSessionsHaveUniqueTokens(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("JWT_SECRET", "test-signing-key-with-at-least-32-characters")
	first, err := GenerateJWT(1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := GenerateJWT(1)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("two sessions received the same token")
	}
}
