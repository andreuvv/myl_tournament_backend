package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates API key for protected routes
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")

		expectedKey := os.Getenv("API_KEY")
		if expectedKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not configured"})
			c.Abort()
			return
		}

		if apiKey != expectedKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing API key"})
			c.Abort()
			return
		}

		c.Next()
	}
}
