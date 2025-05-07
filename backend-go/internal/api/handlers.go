package api

import (
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ran/demo/backend-go/internal/core/models/document"
	ports "github.com/ran/demo/backend-go/internal/core/ports/document"
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

// Health check handler
func (s *Server) handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "document-service",
	})
}

// Upload a document
func (s *Server) handleUploadDocument(c *gin.Context) {
	// Get user ID from context (would come from auth middleware in a real app)
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get space ID from query params
	spaceID := c.Query("space_id")
	if spaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "space_id is required"})
		return
	}

	// Get file from form data
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	// Create document content
	content := &document.DocumentContent{
		OriginalFilename: header.Filename,
		MIMEType:         header.Header.Get("Content-Type"),
		Stream:           file,
		Size:             header.Size,
	}

	// Handle file upload
	metadata, err := s.lifecycle.HandleFileUpload(c.Request.Context(), userID, spaceID, content)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, metadata)
}

// List documents for a user
func (s *Server) handleListDocuments(c *gin.Context) {
	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get query parameters
	var spaceID *string
	if spaceIDParam := c.Query("space_id"); spaceIDParam != "" {
		spaceID = &spaceIDParam
	}

	// Get pagination parameters
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Get documents
	documents, total, err := s.lifecycle.ListUserDocuments(c.Request.Context(), userID, spaceID, offset, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": documents,
		"total": total,
		"pagination": gin.H{
			"offset": offset,
			"limit":  limit,
		},
	})
}

// Get document details
func (s *Server) handleGetDocument(c *gin.Context) {
	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get document ID from path
	docID := document.DocumentID(c.Param("id"))
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}

	// Get document details
	metadata, err := s.lifecycle.GetDocumentDetails(c.Request.Context(), userID, docID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// Get document content
func (s *Server) handleGetDocumentContent(c *gin.Context) {
	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get document ID from path
	docID := document.DocumentID(c.Param("id"))
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}

	// Get format preference
	format := c.DefaultQuery("format", "markdown")

	var result interface{}
	var err error

	switch format {
	case "markdown":
		// Get markdown content
		result, err = s.lifecycle.GetMarkdownContent(c.Request.Context(), userID, docID)
	case "structured":
		// Get full structured content
		result, err = s.lifecycle.GetDocumentContent(c.Request.Context(), userID, docID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported format"})
		return
	}

	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": result,
		"format":  format,
	})
}

// Download document
func (s *Server) handleDownloadDocument(c *gin.Context) {
	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get document ID from path
	docID := document.DocumentID(c.Param("id"))
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}

	// Get format preference
	format := c.DefaultQuery("format", "original")
	var downloadFormat ports.DownloadFormat

	switch format {
	case "original":
		downloadFormat = ports.DownloadFormatOriginal
	case "markdown":
		downloadFormat = ports.DownloadFormatMarkdown
	case "text":
		downloadFormat = ports.DownloadFormatText
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported format"})
		return
	}

	// Handle redirect download if specified
	if c.DefaultQuery("redirect", "false") == "true" {
		// Generate download URL
		downloadInfo, err := s.lifecycle.GetDownloadURL(c.Request.Context(), userID, docID)
		if err != nil {
			handleError(c, err)
			return
		}

		// Redirect to presigned URL
		c.Redirect(http.StatusTemporaryRedirect, downloadInfo.PresignedURL)
		return
	}

	// Get downloadable
	downloadable, err := s.presenter.GetDownloadable(c.Request.Context(), userID, docID, downloadFormat)
	if err != nil {
		handleError(c, err)
		return
	}

	// Set headers for download
	c.Header("Content-Disposition", "attachment; filename="+downloadable.FileName)
	c.Header("Content-Type", downloadable.ContentType)
	c.Header("Content-Length", strconv.FormatInt(downloadable.ContentLength, 10))

	// Stream content
	_, err = io.Copy(c.Writer, downloadable.Content)
	if err != nil {
		s.logger.Error("Failed to stream content", "error", err)
		// Too late to respond with an error at this point
	}
}

// Delete document
func (s *Server) handleDeleteDocument(c *gin.Context) {
	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get document ID from path
	docID := document.DocumentID(c.Param("id"))
	if docID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid document ID"})
		return
	}

	// Check if permanent delete is requested
	permanent := c.DefaultQuery("permanent", "false") == "true"

	// Delete document
	err := s.lifecycle.DeleteDocument(c.Request.Context(), userID, docID, permanent)
	if err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Get presigned upload URL
func (s *Server) handleGetUploadURL(c *gin.Context) {
	// Get user ID from context
	userID := getUserIDFromContext(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Get space ID from query params
	spaceID := c.Query("space_id")
	if spaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "space_id is required"})
		return
	}

	// Parse request body
	var req struct {
		Filename    string `json:"filename" binding:"required"`
		ContentType string `json:"content_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get presigned upload URL
	uploadInfo, err := s.lifecycle.GetPresignedUploadURL(
		c.Request.Context(),
		userID,
		spaceID,
		req.Filename,
		req.ContentType,
	)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, uploadInfo)
}

// Helper function to get user ID from context
// In a real application, this would come from your auth middleware
func getUserIDFromContext(c *gin.Context) string {
	// For demo purposes, we'll use a header or a query parameter
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = c.Query("user_id")
	}

	// For development, use a default user ID if not provided
	if userID == "" {
		userID = "demo-user"
	}

	return userID
}

// Helper function to handle errors
func handleError(c *gin.Context, err error) {
	// Handle domain-specific errors with appropriate status codes
	switch err {
	case document.ErrDocumentNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case document.ErrPermissionDenied, document.ErrStorageAccessDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case document.ErrDocumentProcessing:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case document.ErrContentNotProcessed:
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		// Generic error for all other cases
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
	}
}
