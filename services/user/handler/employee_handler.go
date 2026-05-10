package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"posdigi-user/service"
)

// EmployeeHandler handles HTTP requests for employee operations
type EmployeeHandler struct {
	employeeService *service.EmployeeService
	logger          *logrus.Logger
}

// NewEmployeeHandler creates a new employee handler
func NewEmployeeHandler(employeeService *service.EmployeeService, logger *logrus.Logger) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
		logger:          logger,
	}
}

// CreateEmployee creates a new employee profile
func (h *EmployeeHandler) CreateEmployee(c echo.Context) error {
	var req service.CreateEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	employee, err := h.employeeService.CreateEmployee(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create employee")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success":  true,
		"message":  "Employee profile created successfully",
		"employee": employee,
	})
}

// GetEmployee retrieves an employee by ID
func (h *EmployeeHandler) GetEmployee(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Employee ID is required",
		})
	}

	employee, err := h.employeeService.GetEmployee(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"employee": employee,
	})
}

// GetEmployeeByUserID retrieves an employee by user ID
func (h *EmployeeHandler) GetEmployeeByUserID(c echo.Context) error {
	userID := c.Param("userId")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "User ID is required",
		})
	}

	employee, err := h.employeeService.GetEmployeeByUserID(userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"employee": employee,
	})
}

// GetEmployeeByCode retrieves an employee by employee code
func (h *EmployeeHandler) GetEmployeeByCode(c echo.Context) error {
	code := c.Param("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Employee code is required",
		})
	}

	employee, err := h.employeeService.GetEmployeeByCode(code)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"employee": employee,
	})
}

// ListEmployees retrieves all employees with pagination
func (h *EmployeeHandler) ListEmployees(c echo.Context) error {
	// Get pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	search := c.QueryParam("search")

	employees, total, err := h.employeeService.ListEmployees(page, pageSize, search)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list employees")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to retrieve employees",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":    true,
		"employees":  employees,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	})
}

// UpdateEmployee updates an employee profile
func (h *EmployeeHandler) UpdateEmployee(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Employee ID is required",
		})
	}

	var req service.UpdateEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	employee, err := h.employeeService.UpdateEmployee(id, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update employee")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"message":  "Employee updated successfully",
		"employee": employee,
	})
}

// DeleteEmployee deletes an employee profile
func (h *EmployeeHandler) DeleteEmployee(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Employee ID is required",
		})
	}

	err := h.employeeService.DeleteEmployee(id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete employee")
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Employee deleted successfully",
	})
}

// GetEmployeesByDepartment retrieves employees by department
func (h *EmployeeHandler) GetEmployeesByDepartment(c echo.Context) error {
	department := c.Param("department")
	if department == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Department is required",
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	employees, total, err := h.employeeService.GetEmployeesByDepartment(department, page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get employees by department")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to retrieve employees",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"department": department,
		"employees": employees,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetSubordinates retrieves employees who report to a manager
func (h *EmployeeHandler) GetSubordinates(c echo.Context) error {
	managerID := c.Param("managerId")
	if managerID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Manager ID is required",
		})
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	employees, total, err := h.employeeService.GetSubordinates(managerID, page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get subordinates")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":  true,
		"manager_id": managerID,
		"employees": employees,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetActiveEmployees retrieves all active employees
func (h *EmployeeHandler) GetActiveEmployees(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))

	employees, total, err := h.employeeService.GetActiveEmployees(page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get active employees")
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"message": "Failed to retrieve active employees",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success":   true,
		"employees": employees,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateEmploymentStatus updates the employment status of an employee
func (h *EmployeeHandler) UpdateEmploymentStatus(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Employee ID is required",
		})
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Invalid request format",
		})
	}

	err := h.employeeService.UpdateEmploymentStatus(id, req.Status)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update employment status")
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Employment status updated successfully",
	})
}

// GetEmployeeProfile retrieves complete employee profile
func (h *EmployeeHandler) GetEmployeeProfile(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Employee ID is required",
		})
	}

	profile, err := h.employeeService.GetEmployeeProfile(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"profile":  profile,
	})
}
