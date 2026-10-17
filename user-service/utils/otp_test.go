package utils

import (
	"context"
	"regexp"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func setupOTP(t *testing.T) *miniredis.Miniredis {
	t.Helper()
	srv := miniredis.RunT(t)
	RedisClient = redis.NewClient(&redis.Options{Addr: srv.Addr()})
	t.Setenv("JWT_SECRET", "test-customer-key-with-more-than-32-characters")
	t.Cleanup(func() { RedisClient.Close() })
	return srv
}

func TestOTPExpiryReuseAttemptsAndResend(t *testing.T) {
	srv := setupOTP(t)
	email := "customer@example.test"
	if err := StoreOTP(email, "123456"); err != nil {
		t.Fatal(err)
	}
	if value, _ := RedisClient.Get(context.Background(), "otp:"+email).Result(); value == "123456" {
		t.Fatal("OTP stored in plaintext")
	}
	if !VerifyOTP(email, "123456") || VerifyOTP(email, "123456") {
		t.Fatal("OTP must be accepted exactly once")
	}
	if err := StoreOTP(email, "654321"); err != nil {
		t.Fatal(err)
	}
	srv.FastForward(121 * time.Second)
	if VerifyOTP(email, "654321") {
		t.Fatal("expired OTP accepted")
	}
	if err := StoreOTP(email, "654321"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if VerifyOTP(email, "000000") {
			t.Fatal("incorrect OTP accepted")
		}
	}
	if VerifyOTP(email, "654321") {
		t.Fatal("challenge survives attempt exhaustion")
	}
	if err := StoreOTP(email, "111111"); err != nil {
		t.Fatal(err)
	}
	if !VerifyOTP(email, "111111") {
		t.Fatal("new challenge did not reset attempts")
	}
}

func TestOTPConcurrentConsumption(t *testing.T) {
	setupOTP(t)
	if err := StoreOTP("customer@example.test", "123456"); err != nil {
		t.Fatal(err)
	}
	var accepted atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if VerifyOTP("customer@example.test", "123456") {
				accepted.Add(1)
			}
		}()
	}
	wg.Wait()
	if accepted.Load() != 1 {
		t.Fatalf("accepted %d times", accepted.Load())
	}
}

func TestGenerateOTPFormat(t *testing.T) {
	format := regexp.MustCompile(`^[0-9]{6}$`)
	for i := 0; i < 100; i++ {
		if !format.MatchString(GenerateOTP()) {
			t.Fatal("OTP must have six decimal digits")
		}
	}
}
