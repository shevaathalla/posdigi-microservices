package router

import (
	"posdigi-gateway/config"
	"posdigi-gateway/handler"
	"posdigi-gateway/middleware"
	"posdigi-gateway/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// SetupRoutes configures all routes for the gateway
func SetupRoutes(e *echo.Echo, proxyHandler *handler.ProxyHandler, healthChecker *service.HealthChecker, logger *logrus.Logger) {
	// Get config for JWT secret
	cfg := config.LoadConfig()

	// Public routes (no authentication required)
	public := e.Group("/api/v1")
	public.POST("/auth/register", proxyHandler.ProxyToAuth)
	public.POST("/auth/login", proxyHandler.ProxyToAuth)
	public.POST("/auth/validate", proxyHandler.ProxyToAuth)
	public.GET("/auth/validate", proxyHandler.ProxyToAuth)

	// Protected routes (JWT authentication required)
	protected := e.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret, logger))

	// User service routes
	// NOTE: both the bare path (/users) AND wildcard (/users/*) are needed.
	// /users/* alone does NOT match GET /api/v1/users (the list endpoint).
	protected.Any("/users", proxyHandler.ProxyToUser)
	protected.Any("/users/*", proxyHandler.ProxyToUser)

	// Employee routes (fully proxied to User Service)
	protected.Any("/employees", proxyHandler.ProxyToUser)
	protected.Any("/employees/*", proxyHandler.ProxyToUser)

	// Attendance service routes
	protected.Any("/attendance", proxyHandler.ProxyToAttendance)
	protected.Any("/attendance/*", proxyHandler.ProxyToAttendance)

	// Gateway health check endpoint
	e.GET("/health", func(c echo.Context) error {
		healthStatus := healthChecker.GetHealthStatus()
		allHealthy := healthChecker.IsAllHealthy()

		statusCode := 200
		if !allHealthy {
			statusCode = 503
		}

		return c.JSON(statusCode, map[string]interface{}{
			"success": allHealthy,
			"service": "gateway",
			"status":  "healthy",
			"services": map[string]interface{}{
				"auth":       healthStatus["auth"],
				"user":       healthStatus["user"],
				"attendance": healthStatus["attendance"],
			},
		})
	})

	logger.Info("Routes configured successfully")
}
