package originguard

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func status(allowed []string, method, host, origin string) int {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Middleware(allowed))
	router.Any("/", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	req := httptest.NewRequest(method, "/", nil)
	req.Host = host
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code
}

func TestOriginGuard(t *testing.T) {
	allowed := []string{"https://app.example.com"}
	tests := []struct {
		name                 string
		method, host, origin string
		want                 int
	}{
		{"same origin", "POST", "app.example.com", "https://app.example.com", 200},
		{"allowed origin", "POST", "api.example.com", "https://app.example.com", 200},
		{"foreign origin", "POST", "api.example.com", "https://evil.example.net", 403},
		{"sibling subdomain", "DELETE", "api.example.com", "https://blog.example.com", 403},
		{"no origin header", "POST", "api.example.com", "", 200},
		{"safe method ignored", "GET", "api.example.com", "https://evil.example.net", 200},
	}
	for _, tc := range tests {
		if got := status(allowed, tc.method, tc.host, tc.origin); got != tc.want {
			t.Errorf("%s: expected %d, got %d", tc.name, tc.want, got)
		}
	}
}
