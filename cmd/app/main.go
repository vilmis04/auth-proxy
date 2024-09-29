package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vilmis04/auth-proxy/internal/auth"
	"github.com/vilmis04/auth-proxy/internal/proxy"
)

func loadEnvVars() {
	PORT := os.Getenv("PORT")
	if PORT == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("[Server] Failed to load environment variables")
		}
	}
}

func loadAllowedOrigins() []string {
	ALLOWED_ORIGINS := os.Getenv("ALLOWED_ORIGINS")
	if ALLOWED_ORIGINS == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("[Server] Failed to load environment variables")
		}
	}

	return strings.Split(ALLOWED_ORIGINS, ",")
}

func init() {
	loadEnvVars()
}

func main() {
	server := gin.Default()
	config := cors.DefaultConfig()
	config.AllowCredentials = true
	config.AllowOrigins = loadAllowedOrigins()
	server.Use(cors.New(config))
	apiRoutes := server.Group("api")
	auth.NewController(apiRoutes).Use()
	server.Use(proxy.AuthMiddleware(), proxy.ProxyMiddleware())

	server.Run()
}
