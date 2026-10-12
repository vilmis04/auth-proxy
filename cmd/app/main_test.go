package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vilmis04/auth-proxy/internal/accessToken"
	"github.com/vilmis04/auth-proxy/internal/config"
)

// gin's response writer needs CloseNotify, which httptest.ResponseRecorder lacks.
type recorder struct{ *httptest.ResponseRecorder }

func (recorder) CloseNotify() <-chan bool { return make(chan bool) }

func newRecorder() recorder { return recorder{httptest.NewRecorder()} }

const testKey = "0123456789abcdef0123456789abcdef"

func testConfig(upstream string) *config.Config {
	u, _ := url.Parse(upstream)
	return &config.Config{
		JWTKey:        testKey,
		JWTTTL:        time.Hour,
		ServiceURL:    u,
		InternalToken: "internal-secret",
	}
}

func TestNewServerFailsWithWeakJWTKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for name, key := range map[string]string{"empty": "", "short": "too-short"} {
		cfg := testConfig("http://service:8080")
		cfg.JWTKey = key
		if _, err := newServer(cfg, nil); err == nil {
			t.Errorf("%s key: expected startup to fail", name)
		}
	}
}

func TestStartupFailsWithEmptyEnvironment(t *testing.T) {
	_, err := config.Load(func(string) string { return "" })
	if err == nil || !strings.Contains(err.Error(), "JWT_KEY") {
		t.Errorf("expected config error naming JWT_KEY, got %v", err)
	}
}

// A request through the public entry point reaches the service with the
// user and internal token headers set, and a forged user header is replaced.
func TestRequestReachesServiceWithUserHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var got http.Header
	var gotPath string
	service := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, gotPath = r.Header.Clone(), r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer service.Close()

	server, err := newServer(testConfig(service.URL), nil)
	if err != nil {
		t.Fatal(err)
	}
	signer, _ := accessToken.NewSigner(testKey, time.Hour)
	token, _ := signer.Create("alice")

	req := httptest.NewRequest(http.MethodGet, "/api/things", nil)
	req.Header.Set("User", "admin")
	req.AddCookie(&http.Cookie{Name: accessToken.ACCESS_TOKEN, Value: *token})
	req.AddCookie(&http.Cookie{Name: "theme", Value: "dark"})
	rec := newRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if gotPath != "/api/things" {
		t.Errorf("path should be forwarded unchanged, got %q", gotPath)
	}
	if got.Get("User") != "alice" {
		t.Errorf("expected user header alice, got %q", got.Values("User"))
	}
	if got.Get("X-Internal-Token") != "internal-secret" {
		t.Errorf("missing internal token, got %q", got.Get("X-Internal-Token"))
	}
	if cookie := got.Get("Cookie"); strings.Contains(cookie, accessToken.ACCESS_TOKEN) || !strings.Contains(cookie, "theme=dark") {
		t.Errorf("JWT cookie should be stripped and others kept, got %q", cookie)
	}
}

func TestUnauthenticatedRequestNeverReachesService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	called := false
	service := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	defer service.Close()

	server, _ := newServer(testConfig(service.URL), nil)
	req := httptest.NewRequest(http.MethodGet, "/api/things", nil)
	req.Header.Set("User", "admin")
	rec := newRecorder()
	server.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized || called {
		t.Errorf("expected 401 without reaching the service, got %d (called=%v)", rec.Code, called)
	}
}

func TestHealthzIsPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server, _ := newServer(testConfig("http://service:8080"), nil)
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestCrossOriginPostIsRejected(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server, _ := newServer(testConfig("http://service:8080"), nil)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader("{}"))
	req.Host = "app.example.com"
	req.Header.Set("Origin", "https://evil.example.net")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
