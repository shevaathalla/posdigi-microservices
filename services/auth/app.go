package main

import (
	"os"

	"posdigi-auth/client"
	"posdigi-auth/config"
	"posdigi-auth/database"
	"posdigi-auth/handler"
	"posdigi-auth/repository"
	"posdigi-auth/router"
	"posdigi-auth/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// App represents the application structure
type App struct {
	Config      *config.Config
	Router      *echo.Echo
	Logger      *logrus.Logger
	AuthHandler *handler.AuthHandler
}

// Bootstrap initializes the application
func Bootstrap() (*App, error) {
	// Initialize logger
	log := config.InitLogger()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	db, err := database.InitPostgres(cfg)
	if err != nil {
		return nil, err
	}

	// Auto-migrate database schema
	if err := db.AutoMigrate(&repository.AuthUser{}); err != nil {
		return nil, err
	}

	// Initialize layers
	authRepo := repository.NewAuthRepository(db)

	// Initialize User Service client
	userServiceURL := getEnv("USER_SERVICE_URL", "http://localhost:8002")
	userClient := client.NewUserClient(userServiceURL)

	authService := service.NewAuthService(authRepo, userClient, cfg)
	authHandler := handler.NewAuthHandler(authService)

	// Setup router
	e := router.Setup(log, authHandler)

	return &App{
		Config:      cfg,
		Router:      e,
		Logger:      log,
		AuthHandler: authHandler,
	}, nil
}

// getEnv retrieves environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
