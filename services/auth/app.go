package main

import (
	"os"

	"posdigi-auth/client"
	"posdigi-auth/config"
	"posdigi-auth/database"
	"posdigi-auth/handler"
	"posdigi-auth/router"
	"posdigi-auth/service"

	"github.com/labstack/echo/v4"
	"github.com/shevaathalla/posdigi-microservice/shared/activitylogger"
	"github.com/shevaathalla/posdigi-microservice/shared/mongodb"
	"github.com/sirupsen/logrus"
)

// App represents the application structure
type App struct {
	Config         *config.Config
	Router         *echo.Echo
	Logger         *logrus.Logger
	AuthHandler    *handler.AuthHandler
	MongoClient    *mongodb.Client
	ActivityLogger *activitylogger.Logger
}

// Bootstrap initializes the application
func Bootstrap() (*App, error) {
	// Initialize logger
	log := config.InitLogger()

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database (kept for future use / migrations)
	_, err := database.InitPostgres(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize MongoDB for activity logging
	var mongoClient *mongodb.Client
	var activityLogger *activitylogger.Logger

	if cfg.MongoDB != nil {
		mongoCfg := mongodb.Config{
			Host:           cfg.MongoDB.Host,
			Port:           cfg.MongoDB.Port,
			DatabaseName:   cfg.MongoDB.DatabaseName,
			Username:       cfg.MongoDB.Username,
			Password:       cfg.MongoDB.Password,
			AuthDB:         cfg.MongoDB.AuthDB,
			ConnectTimeout: cfg.MongoDB.ConnectTimeout,
			PoolLimit:      cfg.MongoDB.PoolLimit,
		}

		mongoClient, err = mongodb.ConnectMongoDB(mongoCfg)
		if err != nil {
			log.Warnf("Failed to connect to MongoDB: %v. Activity logging will be disabled.", err)
		} else {
			// Initialize activity logger
			activityRepo := activitylogger.NewRepository(mongoClient.Database())
			activityLogger = activitylogger.NewLogger(activityRepo, activitylogger.ServiceAuth)
			log.Info("MongoDB activity logging initialized")
		}
	}

	// Initialize User Service client
	userServiceURL := getEnv("USER_SERVICE_URL", "http://localhost:8002")
	userClient := client.NewUserClient(userServiceURL)

	// Initialize service and handler with activity logger
	authService := service.NewAuthService(userClient, cfg, activityLogger)
	authHandler := handler.NewAuthHandler(authService)

	// Setup router
	e := router.Setup(log, authHandler)

	// Add activity logging middleware if logger is initialized
	if activityLogger != nil {
		activityMiddleware := activitylogger.NewMiddleware(activityLogger)
		e.Use(activityMiddleware.ActivityLoggingMiddleware())
		log.Info("Activity logging middleware enabled")
	}

	return &App{
		Config:         cfg,
		Router:         e,
		Logger:         log,
		AuthHandler:    authHandler,
		MongoClient:    mongoClient,
		ActivityLogger: activityLogger,
	}, nil
}

// getEnv retrieves environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
