package router

import (
	"posdigi-attendance/handler"
	"posdigi-attendance/middleware"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/sirupsen/logrus"
)

// Setup configures the Echo router with all routes and middleware
func Setup(log *logrus.Logger, attendanceHandler *handler.AttendanceHandler) *echo.Echo {
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
			"service": "attendance-service",
		})
	})

	// Swagger documentation
	e.GET("/docs/*", echoSwagger.WrapHandler)

	// API routes
	setupAPIRoutes(e, attendanceHandler)

	return e
}

// setupAPIRoutes configures API v1 routes
func setupAPIRoutes(e *echo.Echo, attendanceHandler *handler.AttendanceHandler) {
	api := e.Group("/api/v1")
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
