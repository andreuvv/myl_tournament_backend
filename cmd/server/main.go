package main

import (
	"log"
	"os"

	"github.com/andreuvv/premier_mitologico/backend/internal/database"
	"github.com/andreuvv/premier_mitologico/backend/internal/handlers"
	"github.com/andreuvv/premier_mitologico/backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run database migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Set up Gin router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	// Public routes (no authentication required)
	public := router.Group("/api")
	{
		public.GET("/fixture", handlers.GetFixture)
		public.GET("/standings", handlers.GetStandings)
		public.GET("/players", handlers.GetPlayers)
	}

	// Protected routes (require API key)
	protected := router.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// Match score updates
		protected.PATCH("/matches/:id/score", handlers.UpdateMatchScore)

		// Player management
		protected.POST("/players", handlers.CreatePlayer)

		// Fixture creation (creates entire tournament structure)
		protected.POST("/fixture", handlers.CreateFixture)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
