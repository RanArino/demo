package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ran/demo/backend-go/internal/config"
	ports "github.com/ran/demo/backend-go/internal/core/ports/document"
)

// Server represents the HTTP server for the API
type Server struct {
	router     *gin.Engine
	httpServer *http.Server
	config     *config.Config
	logger     *slog.Logger
	lifecycle  ports.DocumentLifecycleManagerPort
	presenter  ports.DocumentPresenterPort
}

// NewServer creates a new API server
func NewServer(
	cfg *config.Config,
	lifecycle ports.DocumentLifecycleManagerPort,
	presenter ports.DocumentPresenterPort,
	logger *slog.Logger,
) *Server {
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(LoggerMiddleware(logger))

	server := &Server{
		router:    router,
		config:    cfg,
		logger:    logger,
		lifecycle: lifecycle,
		presenter: presenter,
	}

	// Initialize routes
	server.registerRoutes()

	// Create HTTP server
	server.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	return server
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("Starting API server", "address", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down API server")
	return s.httpServer.Shutdown(ctx)
}

// registerRoutes sets up all API routes
func (s *Server) registerRoutes() {
	// Health check
	s.router.GET("/health", s.handleHealthCheck)

	// API group
	api := s.router.Group("/api/v1")

	// Document routes
	docs := api.Group("/documents")
	docs.POST("", s.handleUploadDocument)
	docs.GET("", s.handleListDocuments)
	docs.GET("/:id", s.handleGetDocument)
	docs.GET("/:id/content", s.handleGetDocumentContent)
	docs.GET("/:id/download", s.handleDownloadDocument)
	docs.DELETE("/:id", s.handleDeleteDocument)

	// Upload URLs
	docs.POST("/upload-url", s.handleGetUploadURL)
}

// LoggerMiddleware creates a Gin middleware for logging requests
func LoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// Process request
		c.Next()

		// Log after request is processed
		latency := time.Since(start)
		status := c.Writer.Status()

		// Log at appropriate level based on status code
		logLevel := slog.LevelInfo
		if status >= 400 && status < 500 {
			logLevel = slog.LevelWarn
		} else if status >= 500 {
			logLevel = slog.LevelError
		}

		logger.Log(c.Request.Context(), logLevel, "API Request",
			"status", status,
			"method", c.Request.Method,
			"path", path,
			"ip", c.ClientIP(),
			"latency", latency.String(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
