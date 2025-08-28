package routes

import (
	"user-management-api/internal/handlers"
	"user-management-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures authentication related routes
func SetupAuthRoutes(rg *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth := rg.Group("/auth")
	{
		// Apply moderate rate limiting to auth routes to prevent brute force
		auth.POST("/register", 
			middleware.ModerateRateLimit(), 
			middleware.SingleImageUpload(), 
			authHandler.Register,
		)
		auth.POST("/login", middleware.StrictRateLimit(), authHandler.Login)
	}
}