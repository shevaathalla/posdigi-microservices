package router

import (
	"posdigi-user/handler"
	"posdigi-user/middleware"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/sirupsen/logrus"
)

// Setup configures the Echo router with all routes and middleware
func Setup(log *logrus.Logger, userHandler *handler.UserHandler) *echo.Echo {
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
	setupAPIRoutes(e, userHandler)

	return e
}

// setupAPIRoutes configures API v1 routes
func setupAPIRoutes(e *echo.Echo, userHandler *handler.UserHandler) {
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
}
