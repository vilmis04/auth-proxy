package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func newRouter(l *Limiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/", l.Middleware(), func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	return router
}

func do(router *gin.Engine, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = ip + ":1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestLimitsPerIP(t *testing.T) {
	now := time.Now()
	l := New(rate.Every(time.Minute), 2)
	l.now = func() time.Time { return now }
	router := newRouter(l)

	for i := 0; i < 2; i++ {
		if rec := do(router, "10.0.0.1"); rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	rec := do(router, "10.0.0.1")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header")
	}

	if rec := do(router, "10.0.0.2"); rec.Code != http.StatusOK {
		t.Errorf("other IP should be unaffected, got %d", rec.Code)
	}
}

func TestRecoversAfterRefill(t *testing.T) {
	now := time.Now()
	l := New(rate.Every(time.Minute), 1)
	l.now = func() time.Time { return now }
	router := newRouter(l)

	do(router, "10.0.0.1")
	if rec := do(router, "10.0.0.1"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}

	now = now.Add(time.Minute)
	if rec := do(router, "10.0.0.1"); rec.Code != http.StatusOK {
		t.Errorf("expected 200 after refill, got %d", rec.Code)
	}
}

func TestEvictsIdleClients(t *testing.T) {
	now := time.Now()
	l := New(rate.Every(time.Minute), 1)
	l.now = func() time.Time { return now }
	router := newRouter(l)

	do(router, "10.0.0.1")
	now = now.Add(2 * idleTTL)
	do(router, "10.0.0.2")

	if _, ok := l.clients["10.0.0.1"]; ok {
		t.Error("idle client should have been evicted")
	}
}
