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
func SetupRoutes(cfg *config.Config, healthHandler *handlers.HealthHandler, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler) *gin.Engine {
	if cfg.Server.Env == "Production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()

	// Apply global middleware
	router.Use(middleware.LoggingMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(gin.Recovery())

	// Health check endpoint
	router.GET("/health", healthHandler.HealthCheck)

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Setup API routes
	setupAPIRoutes(router, cfg, authHandler, userHandler)

	return router
}

// setupAPIRoutes configures the API v1 routes
func setupAPIRoutes(router *gin.Engine, cfg *config.Config, authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler) {
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		setupAuthRoutes(v1, authHandler)
		
		// User routes
		setupUserRoutes(v1, cfg, userHandler)
	}
}

// setupAuthRoutes configures authentication related routes
func setupAuthRoutes(rg *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}
}

// setupUserRoutes configures user management routes
func setupUserRoutes(rg *gin.RouterGroup, cfg *config.Config, userHandler *handlers.UserHandler) {
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