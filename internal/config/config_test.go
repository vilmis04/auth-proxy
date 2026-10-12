package config

import (
	"strings"
	"testing"
	"time"
)

func validEnv() map[string]string {
	return map[string]string{
		"JWT_KEY":        strings.Repeat("k", 32),
		"DATABASE_URL":   "postgres://u:p@db:5432/auth?sslmode=disable",
		"SERVICE_URL":    "http://service:8080",
		"INTERNAL_TOKEN": "internal-secret",
	}
}

func load(env map[string]string) (*Config, error) {
	return Load(func(key string) string { return env[key] })
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := load(validEnv())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.JWTTTL != 24*time.Hour {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
	if cfg.CookieDomain != "" || len(cfg.AllowedOrigins) != 0 || len(cfg.TrustedProxies) != 0 {
		t.Errorf("optional values should default to empty: %+v", cfg)
	}
	if cfg.ServiceURL.Host != "service:8080" {
		t.Errorf("unexpected service URL %v", cfg.ServiceURL)
	}
}

func TestLoadRequiredVariables(t *testing.T) {
	for _, name := range []string{"JWT_KEY", "DATABASE_URL", "SERVICE_URL", "INTERNAL_TOKEN"} {
		env := validEnv()
		delete(env, name)
		_, err := load(env)
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("missing %s: expected an error naming it, got %v", name, err)
		}
	}
}

func TestLoadReportsAllProblemsTogether(t *testing.T) {
	_, err := load(map[string]string{})
	if err == nil {
		t.Fatal("expected error")
	}
	for _, name := range []string{"JWT_KEY", "DATABASE_URL", "SERVICE_URL", "INTERNAL_TOKEN"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error should mention %s: %v", name, err)
		}
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	cases := map[string]map[string]string{
		"short key":       {"JWT_KEY": "short"},
		"bad port":        {"PORT": "http"},
		"port range":      {"PORT": "70000"},
		"bad ttl":         {"JWT_TTL": "tomorrow"},
		"negative ttl":    {"JWT_TTL": "-1h"},
		"service no host": {"SERVICE_URL": "service"},
		"service scheme":  {"SERVICE_URL": "ftp://service"},
		"bad proxy":       {"TRUSTED_PROXIES": "10.0.0.0/8,not-an-ip"},
	}
	for name, override := range cases {
		env := validEnv()
		for k, v := range override {
			env[k] = v
		}
		if _, err := load(env); err == nil {
			t.Errorf("%s: expected error", name)
		}
	}
}

func TestLoadParsesOptionalValues(t *testing.T) {
	env := validEnv()
	env["PORT"] = "9000"
	env["JWT_TTL"] = "30m"
	env["COOKIE_DOMAIN"] = ".example.com"
	env["ALLOWED_ORIGINS"] = "https://app.example.com, https://admin.example.com ,"
	env["TRUSTED_PROXIES"] = "10.0.0.0/8, 172.18.0.2"

	cfg, err := load(env)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "9000" || cfg.JWTTTL != 30*time.Minute || cfg.CookieDomain != ".example.com" {
		t.Errorf("unexpected config %+v", cfg)
	}
	if len(cfg.AllowedOrigins) != 2 || cfg.AllowedOrigins[1] != "https://admin.example.com" {
		t.Errorf("unexpected origins %v", cfg.AllowedOrigins)
	}
	if len(cfg.TrustedProxies) != 2 {
		t.Errorf("unexpected trusted proxies %v", cfg.TrustedProxies)
	}
}
