package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestInventoryAdminAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, identity   string
		status, upstream int
		token            string
	}{
		{"anonymous", "", 401, 200, ""},
		{"forged header", "", 401, 200, ""},
		{"customer token rejected", "", 401, 401, "Bearer customer"},
		{"viewer rejected", `{"admin":{"id":1,"role":"viewer","is_active":true}}`, 403, 200, "Bearer viewer"},
		{"inactive rejected", `{"admin":{"id":1,"role":"admin","is_active":false}}`, 403, 200, "Bearer inactive"},
		{"admin permitted", `{"admin":{"id":1,"role":"admin","is_active":true}}`, 204, 200, "Bearer admin"},
		{"super admin permitted", `{"admin":{"id":1,"role":"super_admin","is_active":true}}`, 204, 200, "Bearer super"},
		{"upstream failure", "", 503, 500, "Bearer token"},
		{"malformed identity", "not-json", 503, 200, "Bearer token"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			identity := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/admin/me" || r.Header.Get("Authorization") != tc.token {
					t.Error("identity validation request incorrect")
				}
				w.WriteHeader(tc.upstream)
				_, _ = w.Write([]byte(tc.identity))
			}))
			defer identity.Close()
			t.Setenv("ADMIN_SERVICE_URL", identity.URL)
			r := gin.New()
			r.PUT("/price", requireIdentity(true), func(c *gin.Context) { c.Status(204) })
			req := httptest.NewRequest(http.MethodPut, "/price", nil)
			req.Header.Set("Authorization", tc.token)
			req.Header.Set("X-User-Role", "admin")
			result := httptest.NewRecorder()
			r.ServeHTTP(result, req)
			if result.Code != tc.status {
				t.Fatalf("got %d want %d", result.Code, tc.status)
			}
		})
	}
}
