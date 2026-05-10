package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

// Recover returns a middleware which recovers from panics anywhere in the chain
func Recover() echo.MiddlewareFunc {
	return middleware.Recover()
}

// CORS returns a CORS middleware
func CORS() echo.MiddlewareFunc {
	return middleware.CORS()
}

// Gzip returns a Gzip middleware
func Gzip() echo.MiddlewareFunc {
	return middleware.Gzip()
}

// RequestID returns a middleware that adds a request ID to each request
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get("X-Request-ID")
			if reqID == "" {
				reqID = c.Response().Header().Get(echo.HeaderXRequestID)
			}
			if reqID == "" {
				reqID = "req-" + c.RealIP()
			}
			c.Set("requestId", reqID)
			c.Response().Header().Set("X-Request-ID", reqID)
			return next(c)
		}
	}
}

// Logger returns a middleware that logs HTTP requests
func Logger(log *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Get request ID from context
			requestID := c.Get("requestId")
			if requestID == nil {
				requestID = "unknown"
			}

			// Log request
			log.WithFields(logrus.Fields{
				"request_id": requestID,
				"method":     req.Method,
				"path":       req.URL.Path,
				"remote_ip":  c.RealIP(),
				"user_agent": req.UserAgent(),
			}).Info("Incoming request")

			// Process request
			err := next(c)

			// Log response
			log.WithFields(logrus.Fields{
				"request_id": requestID,
				"method":     req.Method,
				"path":       req.URL.Path,
				"status":     res.Status,
			}).Info("Request completed")

			return err
		}
	}
}
