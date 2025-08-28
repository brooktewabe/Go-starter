package routes

import (
	"user-management-api/internal/config"
	"user-management-api/internal/handlers"
	"user-management-api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupUserRoutes configures user management routes
func SetupUserRoutes(rg *gin.RouterGroup, cfg *config.Config, userHandler *handlers.UserHandler) {
	users := rg.Group("/users")
	{
		// Public user routes (require authentication)
		users.GET("/profile", middleware.AuthMidddleware(cfg), userHandler.GetProfile)
		
		// Admin-only user routes (require authentication + admin role)
		users.GET("", middleware.AuthMidddleware(cfg), middleware.RequireRole("admin"), userHandler.ListUsers)
		users.POST("", middleware.AuthMidddleware(cfg), middleware.RequireRole("admin"), userHandler.CreateUser)
		users.GET("/:id", middleware.AuthMidddleware(cfg), middleware.RequireRole("admin"), userHandler.GetUser)
		users.PUT("/:id", middleware.AuthMidddleware(cfg), middleware.RequireRole("admin"), userHandler.UpdateUser)
		users.DELETE("/:id", middleware.AuthMidddleware(cfg), middleware.RequireRole("admin"), userHandler.DeleteUser)
	}
}