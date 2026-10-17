package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Validate through the authoritative service, rather than trusting browser headers
// or sharing customer/admin signing secrets with inventory.
func requireIdentity(admin bool) gin.HandlerFunc {
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	return func(c *gin.Context) {
		bearer := c.GetHeader("Authorization")
		if !strings.HasPrefix(bearer, "Bearer ") || len(strings.TrimSpace(strings.TrimPrefix(bearer, "Bearer "))) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Verified session required"})
			return
		}
		base, path := os.Getenv("USER_SERVICE_URL"), "/api/v1/users/profile"
		if admin {
			base, path = os.Getenv("ADMIN_SERVICE_URL"), "/api/v1/admin/me"
		}
		if base == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Identity service not configured"})
			return
		}
		req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, base+path, nil)
		if err != nil {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		req.Header.Set("Authorization", bearer)
		resp, err := client.Do(req)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "Identity service unavailable"})
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			c.AbortWithStatusJSON(resp.StatusCode, gin.H{"error": "Access denied"})
			return
		}
		if resp.StatusCode != http.StatusOK {
			c.AbortWithStatus(http.StatusServiceUnavailable)
			return
		}
		if admin {
			var result struct {
				Admin struct {
					ID       int    `json:"id"`
					Role     string `json:"role"`
					IsActive bool   `json:"is_active"`
				} `json:"admin"`
			}
			if json.NewDecoder(http.MaxBytesReader(c.Writer, resp.Body, 65536)).Decode(&result) != nil {
				c.AbortWithStatus(http.StatusServiceUnavailable)
				return
			}
			if result.Admin.ID <= 0 || !result.Admin.IsActive || (result.Admin.Role != "admin" && result.Admin.Role != "super_admin") {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Administrator access required"})
				return
			}
			c.Set("admin_id", result.Admin.ID)
		}
		c.Next()
	}
}
