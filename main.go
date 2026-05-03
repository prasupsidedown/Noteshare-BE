package main

import (
	"log"
	"noteshare-be/config"
	"noteshare-be/database"
	"noteshare-be/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load config dari .env
	config.Load()

	// 2. Connect ke database
	database.Connect()

	// 3. Setup Gin
	if config.AppConfig.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 4. CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 5. Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "noteshare-backend",
			"version": "1.0.0",
		})
	})

	// 6. Setup semua routes
	routes.Setup(r)

	// 7. Jalankan server
	log.Printf("🚀 Noteshare backend running on port %s", config.AppConfig.Port)
	if err := r.Run(":" + config.AppConfig.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}