package middleware

import (
	"posdigi-gateway/config"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// SetupMiddleware configures and registers all middleware for the Echo instance
func SetupMiddleware(e *echo.Echo, cfg *config.Config, log *logrus.Logger) {
	// Create rate limiter
	rateLimiter := NewRateLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)

	// Global middleware
	e.Use(Recover())
	e.Use(CORS())
	e.Use(Gzip())
	e.Use(RequestID())
	e.Use(Logger(log))
	e.Use(RateLimitMiddleware(rateLimiter, log))
}
