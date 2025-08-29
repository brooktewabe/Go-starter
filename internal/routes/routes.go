package routes

import (
	"user-management-api/internal/config"
	"user-management-api/internal/handlers"
	"user-management-api/internal/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes configures all the application routes
func SetupRoutes(cfg *config.Config, healthHandler *handlers.HealthHandler, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, fileHandler *handlers.FileHandler) *gin.Engine {
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()

	// Apply global middleware
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", healthHandler.HealthCheck)

	router.Static("/api/v1/uploads", "./uploads")

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup API routes
	setupAPIRoutes(router, cfg, authHandler, userHandler, fileHandler)

	return router
}

// setupAPIRoutes configures the API v1 routes
func setupAPIRoutes(router *gin.Engine, cfg *config.Config, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, fileHandler *handlers.FileHandler) {
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		SetupAuthRoutes(v1, authHandler)
		
		// User routes
		SetupUserRoutes(v1, cfg, userHandler)
		
		// File routes
		SetupFileRoutes(v1, cfg, fileHandler)
	}
}