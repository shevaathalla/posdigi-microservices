package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// InternalServiceAuth validates the X-Service-Auth header for internal service communication
func InternalServiceAuth(expectedServiceKey string, log *logrus.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the service auth header
			serviceKey := c.Request().Header.Get("X-Service-Auth")

			// Check if the service key matches
			if serviceKey != expectedServiceKey {
				log.WithFields(logrus.Fields{
					"path":        c.Path(),
					"remote_ip":   c.RealIP(),
					"service_key": serviceKey,
				}).Warn("Unauthorized service request")

				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"success": false,
					"message": "Unauthorized service request",
				})
			}

			// Valid internal service request
			return next(c)
		}
	}
}
