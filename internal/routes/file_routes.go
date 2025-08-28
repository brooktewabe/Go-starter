package routes

import (
	"user-management-api/internal/config"
	"user-management-api/internal/handlers"
	"user-management-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupFileRoutes configures file upload routes
func SetupFileRoutes(rg *gin.RouterGroup, cfg *config.Config, fileHandler *handlers.FileHandler) {
	files := rg.Group("/files")
	{
		// General file upload with default config and moderate rate limiting
		files.POST("/upload", 
			middleware.AuthMidddleware(cfg),
			middleware.ModerateRateLimit(),
			middleware.FileUploadMiddleware(middleware.DefaultFileUploadConfig()),
			fileHandler.UploadFile,
		)

		// Image upload with strict rate limiting (to prevent spam)
		files.POST("/upload/image",
			middleware.AuthMidddleware(cfg),
			middleware.StrictRateLimit(),
			middleware.SingleImageUpload(),
			fileHandler.UploadImage,
		)

		// Document upload with moderate rate limiting
		files.POST("/upload/document",
			middleware.AuthMidddleware(cfg),
			middleware.ModerateRateLimit(),
			middleware.SingleDocumentUpload(),
			fileHandler.UploadDocument,
		)

		// Multiple images upload (max 5) with strict rate limiting
		files.POST("/upload/images",
			middleware.AuthMidddleware(cfg),
			middleware.StrictRateLimit(),
			middleware.MultipleImageUpload(5),
			fileHandler.UploadFile,
		)
	}
}