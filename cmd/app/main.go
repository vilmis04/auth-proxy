package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vilmis04/auth-proxy/internal/auth"
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
	auth.NewController(server).Use()

	server.Run()
}
