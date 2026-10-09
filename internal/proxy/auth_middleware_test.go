package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

const authTestKey = "0123456789abcdef0123456789abcdef"

func authRouter(t *testing.T) (*gin.Engine, *accessToken.Signer, *string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	signer, err := accessToken.NewSigner(authTestKey, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	var seenUser string
	router := gin.New()
	router.Use(AuthMiddleware(signer))
	router.GET("/", func(ctx *gin.Context) {
		seenUser = ctx.GetString(userContextKey)
		ctx.Status(http.StatusOK)
	})
	return router, signer, &seenUser
}

func request(router *gin.Engine, token string) int {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: accessToken.ACCESS_TOKEN, Value: token})
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec.Code
}

func TestAuthMiddlewareAcceptsValidToken(t *testing.T) {
	router, signer, seenUser := authRouter(t)
	token, _ := signer.Create("a&b")

	if code := request(router, *token); code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
	if *seenUser != "a&b" {
		t.Errorf("expected handler to see user a&b, got %q", *seenUser)
	}
}

func TestAuthMiddlewareRejectsBadRequests(t *testing.T) {
	router, _, seenUser := authRouter(t)

	hs384, _ := jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{
		"sub": "alice",
		"exp": jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(authTestKey))
	noExpiry, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "alice"}).SignedString([]byte(authTestKey))
	expired, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "alice",
		"exp": jwt.NewNumericDate(time.Now().Add(-time.Hour)),
	}).SignedString([]byte(authTestKey))
	wrongKey, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "alice",
		"exp": jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte("another-key-another-key-another-key"))

	tests := map[string]string{
		"no cookie":        "",
		"garbage":          "not-a-jwt",
		"expired":          expired,
		"wrong algorithm":  hs384,
		"missing exp":      noExpiry,
		"signed other key": wrongKey,
	}
	for name, token := range tests {
		if code := request(router, token); code != http.StatusUnauthorized {
			t.Errorf("%s: expected 401, got %d", name, code)
		}
	}
	if *seenUser != "" {
		t.Errorf("handler must not run for rejected requests, saw user %q", *seenUser)
	}
}
