package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter creates and configures a new HTTP router
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Register routes
	router.GET("/health", HealthCheck)

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		// Future endpoints will be added here
		v1.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Welcome to TextViz API v1",
			})
		})
	}

	return router
}
