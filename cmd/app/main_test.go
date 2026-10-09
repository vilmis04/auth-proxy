package main

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

func TestNewServerFailsWithoutJWTKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for name, values := range map[string]map[string]string{
		"empty": {},
		"short": {"JWT_KEY": "too-short"},
	} {
		if _, err := newServer(env(values)); err == nil {
			t.Errorf("%s key: expected startup to fail", name)
		}
	}
}

func TestNewServerRejectsBadTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, err := newServer(env(map[string]string{
		"JWT_KEY": strings.Repeat("k", 32),
		"JWT_TTL": "soon",
	}))
	if err == nil {
		t.Error("expected invalid JWT_TTL to fail startup")
	}
}

func TestNewServerStartsWithValidConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server, err := newServer(env(map[string]string{
		"JWT_KEY":         strings.Repeat("k", 32),
		"ALLOWED_ORIGINS": "http://localhost:3000, http://localhost:3300",
	}))
	if err != nil || server == nil {
		t.Fatalf("expected server, got %v", err)
	}
}
