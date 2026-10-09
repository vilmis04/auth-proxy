package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	NewController(router.Group("api"), newTestService(t, newFakeRepo()), Limits{}).Use()
	return router
}

func post(router *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestTokenOnlyInCookie(t *testing.T) {
	router := newTestRouter(t)
	signUp := `{"username":"a&b","password":"correct horse","repeatPassword":"correct horse"}`
	login := `{"username":"a&b","password":"correct horse"}`

	for path, tc := range map[string]struct {
		body   string
		status int
	}{
		"/api/auth/sign-up": {signUp, http.StatusCreated},
		"/api/auth/login":   {login, http.StatusOK},
	} {
		rec := post(router, path, tc.body)
		if rec.Code != tc.status {
			t.Fatalf("%s: expected %d, got %d (%s)", path, tc.status, rec.Code, rec.Body)
		}
		if strings.Contains(strings.ToLower(rec.Body.String()), "token") {
			t.Errorf("%s: body must not contain a token: %s", path, rec.Body)
		}
		var resp UserResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil || resp.Username != "a&b" {
			t.Errorf("%s: unexpected body %s (%v)", path, rec.Body, err)
		}

		cookies := rec.Result().Cookies()
		if len(cookies) != 1 || cookies[0].Name != "access_token" || cookies[0].Value == "" {
			t.Fatalf("%s: expected access_token cookie, got %v", path, cookies)
		}
		if !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].MaxAge != 3600 {
			t.Errorf("%s: unexpected cookie attributes %+v", path, cookies[0])
		}
	}
}

func TestMalformedBodyIsBadRequest(t *testing.T) {
	router := newTestRouter(t)
	for _, path := range []string{"/api/auth/sign-up", "/api/auth/login"} {
		if rec := post(router, path, "{not json"); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", path, rec.Code)
		}
	}
}

func TestOversizedBodyIsRejected(t *testing.T) {
	rec := post(newTestRouter(t), "/api/auth/login", `{"username":"`+strings.Repeat("a", 5000)+`"}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestLoginRateLimitIsApplied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	blocked := func(ctx *gin.Context) { ctx.AbortWithStatus(http.StatusTooManyRequests) }
	NewController(router.Group("api"), newTestService(t, newFakeRepo()), Limits{Login: blocked}).Use()

	if rec := post(router, "/api/auth/login", `{}`); rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected limiter to run before the handler, got %d", rec.Code)
	}
}
