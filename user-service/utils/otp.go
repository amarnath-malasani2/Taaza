package utils

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		logrus.Fatal("Failed to connect to Redis:", err)
	}
	logrus.Info("Redis connected successfully")
}

func GenerateOTP() string {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		panic("cryptographic OTP generation unavailable")
	}
	return fmt.Sprintf("%06d", value.Int64())
}

func otpDigest(email, otp string) string {
	mac := hmac.New(sha256.New, []byte(os.Getenv("JWT_SECRET")))
	mac.Write([]byte(strings.ToLower(strings.TrimSpace(email)) + ":" + otp))
	return hex.EncodeToString(mac.Sum(nil))
}

func StoreOTP(email, otp string) error {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", strings.ToLower(strings.TrimSpace(email)))
	return RedisClient.Eval(ctx, `
	redis.call('SET', KEYS[1], ARGV[1], 'EX', 120)
	redis.call('DEL', KEYS[2])
	return 1`, []string{key, key + ":attempts"}, otpDigest(email, otp)).Err()
}

func VerifyOTP(email, otp string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("otp:%s", strings.ToLower(strings.TrimSpace(email)))
	result, err := RedisClient.Eval(ctx, `
	local expected=redis.call('GET',KEYS[1])
	if not expected then return 0 end
	if expected==ARGV[1] then
	 redis.call('DEL',KEYS[1],KEYS[2]); return 1
	end
	local attempts=redis.call('INCR',KEYS[2])
	redis.call('EXPIRE',KEYS[2],120)
	if attempts>=5 then redis.call('DEL',KEYS[1]) end
	return 0`, []string{key, key + ":attempts"}, otpDigest(email, otp)).Int()
	return err == nil && result == 1
}
