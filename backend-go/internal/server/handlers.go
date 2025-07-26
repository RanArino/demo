package server

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthCheckResponse represents the response for the health check endpoint
type HealthCheckResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// HealthCheck handles health check requests
func HealthCheck(c *gin.Context) {
	response := HealthCheckResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "0.1.0", // This should come from a version file or build info in a real app
	}
	c.JSON(http.StatusOK, response)
}
