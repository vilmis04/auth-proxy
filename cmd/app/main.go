package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/vilmis04/auth-proxy/internal/accessToken"
	"github.com/vilmis04/auth-proxy/internal/auth"
	"github.com/vilmis04/auth-proxy/internal/config"
	"github.com/vilmis04/auth-proxy/internal/health"
	"github.com/vilmis04/auth-proxy/internal/originguard"
	"github.com/vilmis04/auth-proxy/internal/proxy"
	"github.com/vilmis04/auth-proxy/internal/ratelimit"
	"github.com/vilmis04/auth-proxy/internal/storage"
	"golang.org/x/time/rate"
)

// newServer wires the app from validated configuration. It fails fast on a
// weak JWT key even when called without going through config.Load.
func newServer(cfg *config.Config, db *sql.DB) (*gin.Engine, error) {
	signer, err := accessToken.NewSigner(cfg.JWTKey, cfg.JWTTTL)
	if err != nil {
		return nil, err
	}
	upstream, err := proxy.New(cfg.ServiceURL, cfg.InternalToken, cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}

	server := gin.Default()
	// Gin trusts every proxy by default, which would let clients spoof their IP.
	if err := server.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}

	// CORS is only needed when the client lives on a different origin.
	if len(cfg.AllowedOrigins) > 0 {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowCredentials = true
		corsConfig.AllowOrigins = cfg.AllowedOrigins
		server.Use(cors.New(corsConfig))
	}
	server.Use(originguard.Middleware(cfg.AllowedOrigins))

	health.Register(server, db)

	// Login: 10 per minute per IP, burst 5. Sign-up: 5 per hour per IP, burst 3.
	limits := auth.Limits{
		Login:  ratelimit.New(rate.Every(6*time.Second), 5).Middleware(),
		SignUp: ratelimit.New(rate.Every(12*time.Minute), 3).Middleware(),
	}
	service := auth.NewService(auth.NewRepo(db), signer)
	auth.NewController(server.Group("api"), service, limits, cfg.CookieDomain).Use()

	server.NoRoute(proxy.AuthMiddleware(signer), proxy.Handler(upstream))

	return server, nil
}

func main() {
	// Release mode unless GIN_MODE says otherwise (e.g. debug for local work).
	if os.Getenv(gin.EnvGinMode) == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		log.Fatalf("[Server] invalid configuration:\n%v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := storage.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[Server] database: %v", err)
	}
	defer db.Close()
	if err := storage.Migrate(ctx, db); err != nil {
		log.Fatalf("[Server] migrations: %v", err)
	}

	handler, err := newServer(cfg, db)
	if err != nil {
		log.Fatalf("[Server] %v", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("[Server] shutdown: %v", err)
		}
	}()

	log.Printf("[Server] listening on %v", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("[Server] %v", err)
	}
}
