package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vilmis04/auth-proxy/internal/accessToken"
	"github.com/vilmis04/auth-proxy/internal/auth"
	"github.com/vilmis04/auth-proxy/internal/proxy"
	"github.com/vilmis04/auth-proxy/internal/ratelimit"
	"golang.org/x/time/rate"
)

const defaultTokenTTL = 24 * time.Hour

func splitOrigins(value string) []string {
	var origins []string
	for _, origin := range strings.Split(value, ",") {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins = append(origins, origin)
		}
	}

	return origins
}

func tokenTTL(getenv func(string) string) (time.Duration, error) {
	value := getenv("JWT_TTL")
	if value == "" {
		return defaultTokenTTL, nil
	}
	ttl, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("JWT_TTL: %w", err)
	}

	return ttl, nil
}

// newServer wires the app from environment values. It fails fast on bad
// configuration, in particular a missing or weak JWT_KEY.
func newServer(getenv func(string) string) (*gin.Engine, error) {
	ttl, err := tokenTTL(getenv)
	if err != nil {
		return nil, err
	}
	signer, err := accessToken.NewSigner(getenv("JWT_KEY"), ttl)
	if err != nil {
		return nil, err
	}

	server := gin.Default()
	if origins := splitOrigins(getenv("ALLOWED_ORIGINS")); len(origins) > 0 {
		config := cors.DefaultConfig()
		config.AllowCredentials = true
		config.AllowOrigins = origins
		server.Use(cors.New(config))
	}

	// Login: 10 per minute per IP, burst 5. Sign-up: 5 per hour per IP, burst 3.
	limits := auth.Limits{
		Login:  ratelimit.New(rate.Every(6*time.Second), 5).Middleware(),
		SignUp: ratelimit.New(rate.Every(12*time.Minute), 3).Middleware(),
	}
	service := auth.NewService(auth.NewRepo(), signer)
	auth.NewController(server.Group("api"), service, limits).Use()
	server.Use(proxy.AuthMiddleware(signer), proxy.ProxyMiddleware())

	return server, nil
}

func main() {
	// Local development convenience; variables already set take precedence.
	_ = godotenv.Load()

	server, err := newServer(os.Getenv)
	if err != nil {
		log.Fatalf("[Server] invalid configuration: %v", err)
	}

	server.Run()
}
