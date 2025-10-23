package middleware

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"user-management-api/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"slices"
)

// generateUniqueFilename generates a unique filename to prevent conflicts
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := originalFilename[:len(originalFilename)-len(ext)]
	timestamp := time.Now().Unix()
	id := primitive.NewObjectID().Hex()
	return fmt.Sprintf("%s_%d_%s%s", name, timestamp, id, ext)
}

// FileUploadConfig holds configuration for file upload middleware
type FileUploadConfig struct {
	MaxFileSize   int64    // Maximum file size in bytes
	AllowedTypes  []string // Allowed MIME types
	AllowedExts   []string // Allowed file extensions
	UploadPath    string   // Upload directory path
	FieldName     string   // Form field name for file
	Required      bool     // Whether file is required
	MaxFiles      int      // Maximum number of files (for multiple uploads)
}

// DefaultFileUploadConfig returns a default configuration
func DefaultFileUploadConfig() FileUploadConfig {
	return FileUploadConfig{
		MaxFileSize:  10 << 20, // 10MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/gif", "application/pdf"},
		AllowedExts:  []string{".jpg", ".jpeg", ".png", ".gif", ".pdf"},
		UploadPath:   "./uploads",
		FieldName:    "file",
		Required:     true,
		MaxFiles:     1,
	}
}

// ImageUploadConfig returns configuration for image uploads only
func ImageUploadConfig() FileUploadConfig {
	return FileUploadConfig{
		MaxFileSize:  5 << 20, // 5MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
		AllowedExts:  []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		UploadPath:   "./uploads/images",
		FieldName:    "image",
		Required:     true,
		MaxFiles:     1,
	}
}

// DocumentUploadConfig returns configuration for document uploads
func DocumentUploadConfig() FileUploadConfig {
	return FileUploadConfig{
		MaxFileSize:  20 << 20, // 20MB
		AllowedTypes: []string{"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		AllowedExts:  []string{".pdf", ".doc", ".docx"},
		UploadPath:   "./uploads/documents",
		FieldName:    "document",
		Required:     true,
		MaxFiles:     1,
	}
}

func FileUploadMiddleware(config FileUploadConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse multipart form
		if err := c.Request.ParseMultipartForm(config.MaxFileSize); err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Failed to parse multipart form",
				Error:   "INVALID_FORM_DATA",
			})
			c.Abort()
			return
		}

		// Get file from form
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: "Failed to get multipart form",
				Error:   "INVALID_FORM_DATA",
			})
			c.Abort()
			return
		}

		files := form.File[config.FieldName]
		
		// Check if file is required
		if config.Required && len(files) == 0 {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: fmt.Sprintf("File field '%s' is required", config.FieldName),
				Error:   "FILE_REQUIRED",
			})
			c.Abort()
			return
		}

		// Check maximum number of files
		if len(files) > config.MaxFiles {
			c.JSON(http.StatusBadRequest, models.APIResponse{
				Success: false,
				Message: fmt.Sprintf("Maximum %d files allowed", config.MaxFiles),
				Error:   "TOO_MANY_FILES",
			})
			c.Abort()
			return
		}

		// Validate files
		for _, fileHeader := range files {
			if err := validateFile(fileHeader, config); err != nil {
				c.JSON(http.StatusBadRequest, models.APIResponse{
					Success: false,
					Message: err.Error(),
					Error:   "FILE_VALIDATION_FAILED",
				})
				c.Abort()
				return
			}
		}

		// Save files and store normalized paths in context
		var savedPaths []string
		for _, fileHeader := range files {
			// Create upload directory if it doesn't exist
			if err := os.MkdirAll(config.UploadPath, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, models.APIResponse{
					Success: false,
					Message: "Failed to create upload directory",
					Error:   "DIRECTORY_CREATION_FAILED",
				})
				c.Abort()
				return
			}

			// Generate unique filename
			filename := generateUniqueFilename(fileHeader.Filename)
			path := filepath.Join(config.UploadPath, filename)

			// Save file
			if err := c.SaveUploadedFile(fileHeader, path); err != nil {
				c.JSON(http.StatusInternalServerError, models.APIResponse{
					Success: false,
					Message: "Failed to save file",
					Error:   "FILE_SAVE_FAILED",
				})
				c.Abort()
				return
			}

			// Convert path to forward slashes for URL compatibility
			savedPaths = append(savedPaths, filepath.ToSlash(path))
		}
		c.Set("uploadedFiles", savedPaths)
		c.Set("uploadConfig", config)
		c.Next()
	}
}

// validateFile validates a single file against the configuration
func validateFile(fileHeader *multipart.FileHeader, config FileUploadConfig) error {
	// Check file size
	if fileHeader.Size > config.MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", config.MaxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !contains(config.AllowedExts, ext) {
		return fmt.Errorf("file extension '%s' not allowed. Allowed extensions: %v", ext, config.AllowedExts)
	}

	// Open file to check MIME type
	file, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("failed to open file for validation")
	}
	defer file.Close()

	// Read first 512 bytes to detect content type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read file for validation")
	}

	// Detect content type
	contentType := http.DetectContentType(buffer)
	if !contains(config.AllowedTypes, contentType) {
		return fmt.Errorf("file type '%s' not allowed. Allowed types: %v", contentType, config.AllowedTypes)
	}

	return nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	// for _, s := range slice {
	// 	if s == item {
	// 		return true
	// 	}
	// }
	// return false
	return slices.Contains(slice, item)
}

// SingleImageUpload - Middleware for single image upload
func SingleImageUpload() gin.HandlerFunc {
	return FileUploadMiddleware(ImageUploadConfig())
}

// SingleDocumentUpload - Middleware for single document upload
func SingleDocumentUpload() gin.HandlerFunc {
	return FileUploadMiddleware(DocumentUploadConfig())
}

// MultipleImageUpload - Middleware for multiple image uploads (max 5)
func MultipleImageUpload(maxFiles int) gin.HandlerFunc {
	config := ImageUploadConfig()
	config.MaxFiles = maxFiles
	config.FieldName = "images"
	return FileUploadMiddleware(config)
}