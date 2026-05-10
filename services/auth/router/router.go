package router

import (
	"posdigi-auth/handler"
	"posdigi-auth/middleware"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/sirupsen/logrus"
)

// Setup configures the Echo router with all routes and middleware
func Setup(log *logrus.Logger, authHandler *handler.AuthHandler) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	// No Gzip — this service sits behind the gateway; individual services should
	// not compress since it causes double-encoding when proxied
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
	setupAPIRoutes(e, authHandler)

	return e
}

// setupAPIRoutes configures API v1 routes
func setupAPIRoutes(e *echo.Echo, authHandler *handler.AuthHandler) {
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
}
