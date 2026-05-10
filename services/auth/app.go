package main

import (
	"os"

	"posdigi-auth/config"
	"posdigi-auth/database"
	"posdigi-auth/handler"
	"posdigi-auth/middleware"
	"posdigi-auth/repository"
	"posdigi-auth/service"
	"posdigi-auth/client"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
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

	// Initialize HTTP client for User Service communication
	userServiceURL := getEnv("USER_SERVICE_URL", "http://localhost:8002")
	userClient := client.NewHTTPClient(userServiceURL)

	authService := service.NewAuthService(authRepo, userClient, cfg)
	authHandler := handler.NewAuthHandler(authService)

	// Setup router
	e := setupRouter(cfg, log, authHandler)

	return &App{
		Config:      cfg,
		Router:      e,
		Logger:      log,
		AuthHandler: authHandler,
	}, nil
}

// setupRouter configures the Echo router
func setupRouter(cfg *config.Config, log *logrus.Logger, authHandler *handler.AuthHandler) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.Gzip())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(log))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "auth-service",
		})
	})

	// Swagger documentation
	e.GET("/docs/*", echoSwagger.WrapHandler)

	// API routes
	api := e.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/validate", authHandler.ValidateToken)
			auth.GET("/validate", authHandler.ValidateToken)
		}
	}

	return e
}

// getEnv retrieves environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}