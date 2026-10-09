// Package config reads all runtime settings from environment variables.
package config

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vilmis04/auth-proxy/internal/accessToken"
)

const (
	defaultPort     = "8080"
	defaultTokenTTL = 24 * time.Hour
)

type Config struct {
	Port           string
	JWTKey         string
	JWTTTL         time.Duration
	DatabaseURL    string
	ServiceURL     *url.URL
	InternalToken  string
	AllowedOrigins []string
	CookieDomain   string
	TrustedProxies []string
}

// Load validates every variable and reports all problems together.
func Load(getenv func(string) string) (*Config, error) {
	var errs []error
	required := func(name string) string {
		value := strings.TrimSpace(getenv(name))
		if value == "" {
			errs = append(errs, fmt.Errorf("%s is required", name))
		}
		return value
	}

	cfg := &Config{
		Port:           defaultPort,
		JWTTTL:         defaultTokenTTL,
		JWTKey:         required("JWT_KEY"),
		DatabaseURL:    required("DATABASE_URL"),
		InternalToken:  required("INTERNAL_TOKEN"),
		CookieDomain:   strings.TrimSpace(getenv("COOKIE_DOMAIN")),
		AllowedOrigins: splitList(getenv("ALLOWED_ORIGINS")),
		TrustedProxies: splitList(getenv("TRUSTED_PROXIES")),
	}

	if cfg.JWTKey != "" && len(cfg.JWTKey) < accessToken.MinKeyLength {
		errs = append(errs, fmt.Errorf("JWT_KEY must be at least %d bytes", accessToken.MinKeyLength))
	}

	if port := strings.TrimSpace(getenv("PORT")); port != "" {
		if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
			errs = append(errs, fmt.Errorf("PORT must be a number between 1 and 65535, got %q", port))
		} else {
			cfg.Port = port
		}
	}

	if value := strings.TrimSpace(getenv("JWT_TTL")); value != "" {
		ttl, err := time.ParseDuration(value)
		if err != nil || ttl <= 0 {
			errs = append(errs, fmt.Errorf("JWT_TTL must be a positive duration such as 24h, got %q", value))
		} else {
			cfg.JWTTTL = ttl
		}
	}

	if raw := required("SERVICE_URL"); raw != "" {
		u, err := url.Parse(raw)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			errs = append(errs, fmt.Errorf("SERVICE_URL must be an http(s) URL with a host, got %q", raw))
		} else {
			cfg.ServiceURL = u
		}
	}

	for _, proxy := range cfg.TrustedProxies {
		if _, err := netip.ParsePrefix(proxy); err != nil {
			if _, err := netip.ParseAddr(proxy); err != nil {
				errs = append(errs, fmt.Errorf("TRUSTED_PROXIES: %q is not an IP or CIDR", proxy))
			}
		}
	}

	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	return cfg, nil
}

func splitList(value string) []string {
	var items []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}

	return items
}
