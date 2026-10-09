package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func get(path string) int {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	Register(router, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec.Code
}

func TestHealthzDoesNotNeedDatabase(t *testing.T) {
	if code := get("/healthz"); code != http.StatusOK {
		t.Errorf("expected 200, got %d", code)
	}
}

func TestReadyzFailsWithoutDatabase(t *testing.T) {
	if code := get("/readyz"); code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", code)
	}
}
