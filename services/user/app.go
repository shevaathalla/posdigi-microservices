package main

import (
	"posdigi-user/config"
	"posdigi-user/database"
	"posdigi-user/handler"
	"posdigi-user/middleware"
	"posdigi-user/repository"
	"posdigi-user/service"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/sirupsen/logrus"
)

// App represents the application structure
type App struct {
	Config      *config.Config
	Router      *echo.Echo
	Logger      *logrus.Logger
	UserHandler *handler.UserHandler
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
	if err := db.AutoMigrate(&repository.User{}); err != nil {
		return nil, err
	}

	// Initialize layers
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg)
	userHandler := handler.NewUserHandler(userService)

	// Setup router
	e := setupRouter(log, userHandler)

	return &App{
		Config:      cfg,
		Router:      e,
		Logger:      log,
		UserHandler: userHandler,
	}, nil
}

// setupRouter configures the Echo router
func setupRouter(log *logrus.Logger, userHandler *handler.UserHandler) *echo.Echo {
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
			"service": "user-service",
		})
	})

	// Swagger documentation
	e.GET("/docs/*", echoSwagger.WrapHandler)

	// API routes
	api := e.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUserByID)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	return e
}