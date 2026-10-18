package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Security headers middleware
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate nonce for CSP
		nonce := generateNonce()
		c.Set("nonce", nonce)

		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'nonce-" + nonce + "' https://cdn.tailwindcss.com; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'"
		c.Header("Content-Security-Policy", csp)

		// HSTS for HTTPS
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// Rate limiting middleware
func RateLimit() gin.HandlerFunc {
	type window struct {
		start time.Time
		count int
	}
	var mu sync.Mutex
	requests := make(map[string]window)
	return func(c *gin.Context) {
		now := time.Now()
		ip := c.ClientIP()
		mu.Lock()
		for key, item := range requests {
			if now.Sub(item.start) >= time.Minute {
				delete(requests, key)
			}
		}
		item := requests[ip]
		if item.start.IsZero() {
			item.start = now
		}
		limited := item.count >= 120 || (len(requests) >= 10000 && item.count == 0)
		if !limited {
			item.count++
			requests[ip] = item
		}
		mu.Unlock()
		if limited {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// Input validation middleware
func ValidateInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for common attack patterns
		userAgent := c.GetHeader("User-Agent")
		if containsSuspiciousPatterns(userAgent) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			c.Abort()
			return
		}

		// Validate content length
		if c.Request.ContentLength > 1024*1024 { // 1MB limit
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request too large"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func generateNonce() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func containsSuspiciousPatterns(input string) bool {
	suspiciousPatterns := []string{
		"<script", "javascript:", "onload=", "onerror=",
		"eval(", "alert(", "document.cookie",
		"union select", "drop table", "insert into",
	}

	lowerInput := strings.ToLower(input)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}
