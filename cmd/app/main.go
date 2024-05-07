package main

import (
	"log"
	"os"

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

func init() {
	loadEnvVars()
}

func main() {
	server := gin.Default()
	config := cors.DefaultConfig()
	config.AllowCredentials = true
	// TODO: add allowed origin to env var
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:3300", "https://voteforthewinners.eu"}
	server.Use(cors.New(config))
	apiRoutes := server.Group("api")
	auth.NewController(apiRoutes).Use()
	server.Use(proxy.AuthMiddleware(), proxy.ProxyMiddleware())

	server.Run()
}
