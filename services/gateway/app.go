package main

import (
	"posdigi-gateway/client"
	"posdigi-gateway/config"
	"posdigi-gateway/handler"
	"posdigi-gateway/middleware"
	"posdigi-gateway/router"
	"posdigi-gateway/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type App struct {
	Config    *config.Config
	Logger    *logrus.Logger
	Router    *echo.Echo
	AuthClient     *client.ServiceClient
	UserClient     *client.ServiceClient
	AttendanceClient *client.ServiceClient
	HealthChecker  *service.HealthChecker
}

// Bootstrap initializes the application and its dependencies
func Bootstrap() (*App, error) {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	config.InitLogger()
	logger := config.GetLogger()

	logger.Info("Bootstrapping Gateway Service...")

	// Initialize Echo router
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Initialize service clients
	authClient := client.NewServiceClient(cfg.AuthServiceURL, cfg.InternalServiceKey, logger)
	userClient := client.NewServiceClient(cfg.UserServiceURL, cfg.InternalServiceKey, logger)
	attendanceClient := client.NewServiceClient(cfg.AttendanceServiceURL, cfg.InternalServiceKey, logger)

	// Initialize health checker
	healthChecker := service.NewHealthChecker(cfg, logger)
	healthChecker.Start()

	// Initialize proxy handler with service clients
	proxyHandler := handler.NewProxyHandler(authClient, userClient, attendanceClient, logger)

	// Setup middleware
	middleware.SetupMiddleware(e, cfg, logger)

	// Setup routes
	router.SetupRoutes(e, proxyHandler, healthChecker, logger)

	app := &App{
		Config:         cfg,
		Logger:         logger,
		Router:         e,
		AuthClient:     authClient,
		UserClient:     userClient,
		AttendanceClient: attendanceClient,
		HealthChecker:  healthChecker,
	}

	logger.Info("Gateway Service bootstrapped successfully")

	return app, nil
}
