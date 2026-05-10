package handler

import (
	"net/http"
	"strconv"
	"strings"

	"posdigi-user/config"
	"posdigi-user/dto"
	"posdigi-user/service"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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
	var req dto.CreateEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid request format"))
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	employee, err := h.employeeService.CreateEmployee(&req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create employee")
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("User not found"))
		}
		if strings.Contains(errMsg, "already exists") {
			return c.JSON(http.StatusConflict, dto.NewErrorResponse("Employee profile already exists for this user"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to create employee profile"))
	}

	config.Info("Employee profile created: " + req.UserID)
	return c.JSON(http.StatusCreated, dto.NewSuccessResponse("Employee profile created successfully", employee))
}

// GetEmployee retrieves an employee by ID
func (h *EmployeeHandler) GetEmployee(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Employee ID is required"))
	}

	employee, err := h.employeeService.GetEmployee(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("Employee not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to retrieve employee"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Employee retrieved successfully", employee))
}

// GetEmployeeByUserID retrieves an employee by user ID
func (h *EmployeeHandler) GetEmployeeByUserID(c echo.Context) error {
	userID := c.Param("userId")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("User ID is required"))
	}

	employee, err := h.employeeService.GetEmployeeByUserID(userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("Employee not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to retrieve employee"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Employee retrieved successfully", employee))
}

// GetEmployeeByCode retrieves an employee by employee code
func (h *EmployeeHandler) GetEmployeeByCode(c echo.Context) error {
	code := c.Param("code")
	if code == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Employee code is required"))
	}

	employee, err := h.employeeService.GetEmployeeByCode(code)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("Employee not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to retrieve employee"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Employee retrieved successfully", employee))
}

// ListEmployees retrieves all employees with pagination
func (h *EmployeeHandler) ListEmployees(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	search := c.QueryParam("search")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	employees, total, err := h.employeeService.ListEmployees(page, pageSize, search)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list employees")
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to retrieve employees"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Employees retrieved successfully", map[string]interface{}{
		"employees": employees,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}))
}

// UpdateEmployee updates an employee profile
func (h *EmployeeHandler) UpdateEmployee(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Employee ID is required"))
	}

	var req dto.UpdateEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid request format"))
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	employee, err := h.employeeService.UpdateEmployee(id, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update employee")
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("Employee not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to update employee"))
	}

	config.Info("Employee updated: " + id)
	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Employee updated successfully", employee))
}

// DeleteEmployee deletes an employee profile
func (h *EmployeeHandler) DeleteEmployee(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Employee ID is required"))
	}

	err := h.employeeService.DeleteEmployee(id)
	if err != nil {
		h.logger.WithError(err).Error("Failed to delete employee")
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("Employee not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to delete employee"))
	}

	config.Info("Employee deleted: " + id)
	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Employee deleted successfully", nil))
}

// GetEmployeesByDepartment retrieves employees by department
func (h *EmployeeHandler) GetEmployeesByDepartment(c echo.Context) error {
	department := c.Param("department")
	if department == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Department is required"))
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	employees, total, err := h.employeeService.GetEmployeesByDepartment(department, page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get employees by department")
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to retrieve employees"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Employees retrieved successfully", map[string]interface{}{
		"department": department,
		"employees":  employees,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	}))
}

// GetSubordinates retrieves employees who report to a manager
func (h *EmployeeHandler) GetSubordinates(c echo.Context) error {
	managerID := c.Param("managerId")
	if managerID == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Manager ID is required"))
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	employees, total, err := h.employeeService.GetSubordinates(managerID, page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get subordinates")
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to retrieve subordinates"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Subordinates retrieved successfully", map[string]interface{}{
		"manager_id": managerID,
		"employees":  employees,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
	}))
}

// GetActiveEmployees retrieves all active employees
func (h *EmployeeHandler) GetActiveEmployees(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	pageSize, _ := strconv.Atoi(c.QueryParam("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	employees, total, err := h.employeeService.GetActiveEmployees(page, pageSize)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get active employees")
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to retrieve active employees"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Active employees retrieved successfully", map[string]interface{}{
		"employees": employees,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}))
}

// UpdateEmploymentStatus updates the employment status of an employee
func (h *EmployeeHandler) UpdateEmploymentStatus(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Employee ID is required"))
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.Bind(&req); err != nil || req.Status == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Status is required"))
	}

	err := h.employeeService.UpdateEmploymentStatus(id, req.Status)
	if err != nil {
		h.logger.WithError(err).Error("Failed to update employment status")
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("Employee not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to update employment status"))
	}

	config.Info("Employment status updated for employee: " + id)
	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Employment status updated successfully", nil))
}

// GetEmployeeProfile retrieves complete employee profile
func (h *EmployeeHandler) GetEmployeeProfile(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Employee ID is required"))
	}

	profile, err := h.employeeService.GetEmployeeProfile(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("Employee not found"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to retrieve employee profile"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Employee profile retrieved successfully", profile))
}
