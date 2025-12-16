package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/Ponloe/cinemesh-core/internal/database"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not loaded, continuing with environment variables")
	}

	if m := os.Getenv("GIN_MODE"); m != "" {
		gin.SetMode(m)
	}

	if err := database.Connect(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// run migrations to create tables
	if err := database.Migrate(); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	// seeding controlled by SEED_DATA env var
	if err := database.Seed(); err != nil {
		log.Fatalf("seeding failed: %v", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
