package router

import (
	"posdigi-user/handler"
	"posdigi-user/middleware"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"github.com/sirupsen/logrus"
)

// Setup configures the Echo router with all routes and middleware
func Setup(log *logrus.Logger, userHandler *handler.UserHandler, employeeHandler *handler.EmployeeHandler, internalServiceKey string) *echo.Echo {
	e := echo.New()

	// Essential middleware only (no CORS/Gzip - Gateway handles those)
	e.Use(middleware.Recover())
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

	// API routes with internal service authentication
	setupAPIRoutes(e, userHandler, employeeHandler, internalServiceKey, log)

	return e
}

// setupAPIRoutes configures API v1 routes
func setupAPIRoutes(e *echo.Echo, userHandler *handler.UserHandler, employeeHandler *handler.EmployeeHandler, internalServiceKey string, log *logrus.Logger) {
	api := e.Group("/api/v1")
	// Apply internal service authentication to all API routes
	api.Use(middleware.InternalServiceAuth(internalServiceKey, log))
	{
		// User management routes
		users := api.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			// Static sub-paths MUST come before /:id to avoid route conflicts
			users.GET("/email/:email", userHandler.GetUserByEmail)
			users.POST("/authenticate", userHandler.AuthenticateUser)
			// Wildcard param routes last
			users.GET("/:id", userHandler.GetUserByID)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Employee management routes
		employees := api.Group("/employees")
		{
			employees.POST("", employeeHandler.CreateEmployee)
			employees.GET("", employeeHandler.ListEmployees)
			employees.GET("/active", employeeHandler.GetActiveEmployees)
			employees.GET("/:id", employeeHandler.GetEmployee)
			employees.GET("/:id/profile", employeeHandler.GetEmployeeProfile)
			employees.PUT("/:id", employeeHandler.UpdateEmployee)
			employees.DELETE("/:id", employeeHandler.DeleteEmployee)
			employees.PATCH("/:id/status", employeeHandler.UpdateEmploymentStatus)

			// User relationship routes
			employees.GET("/user/:userId", employeeHandler.GetEmployeeByUserID)

			// Department routes
			employees.GET("/department/:department", employeeHandler.GetEmployeesByDepartment)

			// Hierarchy routes
			employees.GET("/code/:code", employeeHandler.GetEmployeeByCode)
			employees.GET("/manager/:managerId/subordinates", employeeHandler.GetSubordinates)
		}
	}
}
