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
	// NOTE: No Gzip here — the gateway is a transparent reverse proxy.
	// Adding Gzip corrupts forwarded response bodies because WriteHeader is
	// called before the middleware can inject Content-Encoding: gzip.
	e.Use(RequestID())
	e.Use(Logger(log))
	e.Use(RateLimitMiddleware(rateLimiter, log))
}
