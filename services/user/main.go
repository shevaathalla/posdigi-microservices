package main

import (
	"posdigi-user/config"

	"github.com/joho/godotenv"
)

// @title User Service API
// @version 1.0
// @description User management service for user profiles and CRUD operations
// @host localhost:8002
// @BasePath /
func main() {
	// Load environment variables
	loadEnv()

	// Bootstrap application
	app, err := Bootstrap()
	if err != nil {
		config.GetLogger().Fatalf("Failed to bootstrap application: %v", err)
	}

	// Start server
	app.Logger.Infof("User Service starting on port %s", app.getPort())
	if err := app.Router.Start(":" + app.getPort()); err != nil {
		app.Logger.Fatalf("Failed to start server: %v", err)
	}
}

// loadEnv loads environment variables from .env file
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		if godotenv.Load("../../.env") != nil {
			// No .env file found, will use system env vars or defaults
			return
		}
	}
}

// getPort returns the port to run the server on
func (app *App) getPort() string {
	port := app.Config.Port
	if port == "" {
		return "8002"
	}
	return port
}
