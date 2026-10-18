package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestConcurrentRateLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RateLimit())
	r.GET("/", func(c *gin.Context) { c.Status(204) })
	var wg sync.WaitGroup
	var allowed, limited atomic.Int32
	for i := 0; i < 150; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = "192.0.2.1:1234"
			res := httptest.NewRecorder()
			r.ServeHTTP(res, req)
			if res.Code == 204 {
				allowed.Add(1)
			} else if res.Code == 429 {
				limited.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 120 || limited.Load() != 30 {
		t.Fatalf("allowed=%d limited=%d", allowed.Load(), limited.Load())
	}
}
