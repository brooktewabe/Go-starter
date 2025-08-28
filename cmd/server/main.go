package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-management-api/internal/config"
	"user-management-api/internal/handlers"
	"user-management-api/internal/repository/mongo"
	"user-management-api/internal/routes"
	"user-management-api/internal/services"
	"user-management-api/pkg/database"

	_ "user-management-api/docs" // This line is needed for swagger

)

// @title Go Gin Layered Architecture API
// @version 1.0
// @description A template for building RESTful APIs in Go using the Gin framework with a layered architecture.
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration", err)
	}

	mongoDb, err := database.NewMongoDB(cfg.Database.URI, cfg.Database.Name, cfg.Database.Timeout)
	if err != nil {
		log.Fatal("failed to connect to mongodb")
	}
	defer mongoDb.Close(context.Background())
	// initialize repositories
	userRepo := mongo.NewUserRepository(mongoDb.Database)

	// initialize services
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpiresIn.String())
	userService := services.NewUserService(userRepo)

	// initialize handler

	healthHandler := handlers.NewHealthHandler()
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	fileHandler := handlers.NewFileHandler()

	// setup router
	router := routes.SetupRoutes(cfg, healthHandler, authHandler, userHandler, fileHandler)

	// start server
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.Server.Port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server forced to shutdown ", err)
	}
	log.Println("server exited")

}
