package main

import (
	"posdigi-user/config"
	"posdigi-user/database"
	"posdigi-user/handler"
	"posdigi-user/repository"
	"posdigi-user/router"
	"posdigi-user/service"

	"github.com/labstack/echo/v4"
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
	e := router.Setup(log, userHandler)

	return &App{
		Config:      cfg,
		Router:      e,
		Logger:      log,
		UserHandler: userHandler,
	}, nil
}
