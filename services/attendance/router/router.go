package router

import (
	"posdigi-attendance/handler"
	"posdigi-attendance/middleware"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/sirupsen/logrus"
)

// Setup configures the Echo router with all routes and middleware
func Setup(log *logrus.Logger, attendanceHandler *handler.AttendanceHandler, internalServiceKey string) *echo.Echo {
	e := echo.New()

	// Essential middleware only (no CORS/Gzip - Gateway handles those)
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(log))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "attendance-service",
		})
	})

	// Swagger documentation
	e.GET("/docs/*", echoSwagger.WrapHandler)

	// API routes with internal service authentication
	setupAPIRoutes(e, attendanceHandler, internalServiceKey, log)

	return e
}

// setupAPIRoutes configures API v1 routes
func setupAPIRoutes(e *echo.Echo, attendanceHandler *handler.AttendanceHandler, internalServiceKey string, log *logrus.Logger) {
	api := e.Group("/api/v1")
	// Apply internal service authentication to all API routes
	api.Use(middleware.InternalServiceAuth(internalServiceKey, log))
	{
		attendance := api.Group("/attendance")
		{
			attendance.POST("/clock-in", attendanceHandler.ClockIn)
			attendance.POST("/clock-out", attendanceHandler.ClockOut)
			attendance.GET("/history/:userId", attendanceHandler.GetAttendanceHistory)
			attendance.GET("/summary/:userId", attendanceHandler.GetAttendanceSummary)
		}
	}
}
