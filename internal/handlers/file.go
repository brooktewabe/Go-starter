package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"user-management-api/internal/middleware"
	"user-management-api/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileHandler struct{}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

// UploadFile godoc
// @Summary      Upload a file
// @Description  Upload a single file with validation
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "File to upload"
// @Security     BearerAuth
// @Success      200  {object}  models.APIResponse{data=map[string]interface{}} "File uploaded successfully"
// @Failure      400  {object}  models.APIResponse "Invalid file or validation failed"
// @Failure      500  {object}  models.APIResponse "Internal server error"
// @Router       /files/upload [post]
func (h *FileHandler) UploadFile(c *gin.Context) {
	// Get uploaded files from context (set by middleware)
	uploadedFiles, exists := c.Get("uploadedFiles")
	if !exists {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "No files found",
			Error:   "NO_FILES",
		})
		return
	}

	files := uploadedFiles.([]*multipart.FileHeader)
	config, _ := c.Get("uploadConfig")
	uploadConfig := config.(middleware.FileUploadConfig)

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadConfig.UploadPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create upload directory",
			Error:   "DIRECTORY_CREATION_FAILED",
		})
		return
	}

	var uploadedFileInfos []map[string]interface{}

	// Process each file
	for _, fileHeader := range files {
		// Generate unique filename
		filename := generateUniqueFilename(fileHeader.Filename)
		filepath := filepath.Join(uploadConfig.UploadPath, filename)

		// Save file
		if err := c.SaveUploadedFile(fileHeader, filepath); err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Message: "Failed to save file",
				Error:   "FILE_SAVE_FAILED",
			})
			return
		}

		// Create file info
		fileInfo := map[string]interface{}{
			"original_name": fileHeader.Filename,
			"filename":      filename,
			"size":          fileHeader.Size,
			"path":          filepath,
			"uploaded_at":   time.Now(),
		}

		uploadedFileInfos = append(uploadedFileInfos, fileInfo)
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "File(s) uploaded successfully",
		Data: map[string]interface{}{
			"files": uploadedFileInfos,
			"count": len(uploadedFileInfos),
		},
	})
}

// UploadImage godoc
// @Summary      Upload an image
// @Description  Upload a single image file
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "Image file to upload"
// @Security     BearerAuth
// @Success      200  {object}  models.APIResponse{data=map[string]interface{}} "Image uploaded successfully"
// @Failure      400  {object}  models.APIResponse "Invalid image or validation failed"
// @Failure      500  {object}  models.APIResponse "Internal server error"
// @Router       /files/upload/image [post]
func (h *FileHandler) UploadImage(c *gin.Context) {
	h.UploadFile(c) // Reuse the same logic
}

// UploadDocument godoc
// @Summary      Upload a document
// @Description  Upload a single document file
// @Tags         files
// @Accept       multipart/form-data
// @Produce      json
// @Param        document  formData  file  true  "Document file to upload"
// @Security     BearerAuth
// @Success      200  {object}  models.APIResponse{data=map[string]interface{}} "Document uploaded successfully"
// @Failure      400  {object}  models.APIResponse "Invalid document or validation failed"
// @Failure      500  {object}  models.APIResponse "Internal server error"
// @Router       /files/upload/document [post]
func (h *FileHandler) UploadDocument(c *gin.Context) {
	h.UploadFile(c) // Reuse the same logic
}

// generateUniqueFilename generates a unique filename to prevent conflicts
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := originalFilename[:len(originalFilename)-len(ext)]
	timestamp := time.Now().Unix()
	id := primitive.NewObjectID().Hex()
	return fmt.Sprintf("%s_%d_%s%s", name, timestamp, id, ext)
}