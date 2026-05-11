package handler

import (
	"net/http"
	"strconv"
	"time"

	"posdigi-attendance/config"
	"posdigi-attendance/dto"
	"posdigi-attendance/service"

	"github.com/labstack/echo/v4"
)

type AttendanceHandler struct {
	attendanceService service.AttendanceService
}

// NewAttendanceHandler creates a new attendance handler instance
func NewAttendanceHandler(attendanceService service.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{
		attendanceService: attendanceService,
	}
}

// ClockIn handles clock-in requests
// @Summary Clock in
// @Description Record a clock-in for a user
// @Tags attendance
// @Accept json
// @Produce json
// @Param request body dto.ClockInRequest true "Clock in details"
// @Success 201 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 409 {object} dto.APIResponse
// @Router /attendance/clock-in [post]
func (h *AttendanceHandler) ClockIn(c echo.Context) error {
	var req dto.ClockInRequest
	if err := c.Bind(&req); err != nil {
		config.Warn("Invalid request body for clock-in")
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	attendance, err := h.attendanceService.ClockIn(&req)
	if err != nil {
		if err.Error() == "user already has an active clock-in" {
			return c.JSON(http.StatusConflict, dto.NewErrorResponse("User already has an active clock-in"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	attendanceResponse := dto.NewAttendanceResponse(
		attendance.ID,
		attendance.UserID,
		attendance.ClockIn,
		attendance.ClockOut,
		attendance.Notes,
		attendance.CreatedAt.Format(time.RFC3339),
		attendance.UpdatedAt.Format(time.RFC3339),
	)

	config.Info("Clock-in successful for user: " + req.UserID)
	return c.JSON(http.StatusCreated, dto.NewSuccessResponse("Clocked in successfully", attendanceResponse))
}

// ClockOut handles clock-out requests
// @Summary Clock out
// @Description Record a clock-out for an attendance
// @Tags attendance
// @Accept json
// @Produce json
// @Param request body dto.ClockOutRequest true "Clock out details"
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Failure 404 {object} dto.APIResponse
// @Router /attendance/clock-out [post]
func (h *AttendanceHandler) ClockOut(c echo.Context) error {
	var req dto.ClockOutRequest
	if err := c.Bind(&req); err != nil {
		config.Warn("Invalid request body for clock-out")
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid request body"))
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	attendance, err := h.attendanceService.ClockOut(&req)
	if err != nil {
		if err.Error() == "attendance not found" {
			return c.JSON(http.StatusNotFound, dto.NewErrorResponse("Attendance not found"))
		}
		if err.Error() == "attendance already clocked out" {
			return c.JSON(http.StatusConflict, dto.NewErrorResponse("Attendance already clocked out"))
		}
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	attendanceResponse := dto.NewAttendanceResponse(
		attendance.ID,
		attendance.UserID,
		attendance.ClockIn,
		attendance.ClockOut,
		attendance.Notes,
		attendance.CreatedAt.Format(time.RFC3339),
		attendance.UpdatedAt.Format(time.RFC3339),
	)

	config.Info("Clock-out successful for attendance: " + req.AttendanceID)
	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Clocked out successfully", attendanceResponse))
}

// GetAttendanceHistory handles retrieving attendance history for a user
// @Summary Get attendance history
// @Description Get attendance history for a user with pagination
// @Tags attendance
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Router /attendance/history/{userId} [get]
func (h *AttendanceHandler) GetAttendanceHistory(c echo.Context) error {
	userID := c.Param("userId")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("User ID is required"))
	}

	// Parse and normalize pagination params
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	req := &dto.GetAttendanceHistoryRequest{
		UserID: userID,
		Page:   page,
		Limit:  limit,
	}
	if err := c.Validate(req); err != nil {
		return err
	}

	response, err := h.attendanceService.GetAttendanceHistory(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Attendance history retrieved successfully", response))
}

// GetAttendanceSummary handles retrieving attendance summary for a user
// @Summary Get attendance summary
// @Description Get attendance summary for a user within a date range
// @Tags attendance
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param start_date query string true "Start date (YYYY-MM-DD)" example(2024-01-01)
// @Param end_date query string true "End date (YYYY-MM-DD)" example(2024-12-31)
// @Success 200 {object} dto.APIResponse
// @Failure 400 {object} dto.APIResponse
// @Router /attendance/summary/{userId} [get]
func (h *AttendanceHandler) GetAttendanceSummary(c echo.Context) error {
	userID := c.Param("userId")
	if userID == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("User ID is required"))
	}

	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	if startDate == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("start_date query parameter is required (format: YYYY-MM-DD, example: 2024-01-01)"))
	}
	if endDate == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("end_date query parameter is required (format: YYYY-MM-DD, example: 2024-12-31)"))
	}

	req := &dto.GetAttendanceSummaryRequest{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Manual validation since we're using query params instead of JSON body
	if req.UserID == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("User ID is required"))
	}
	if req.StartDate == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("start_date query parameter is required (format: YYYY-MM-DD, example: 2024-01-01)"))
	}
	if req.EndDate == "" {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("end_date query parameter is required (format: YYYY-MM-DD, example: 2024-12-31)"))
	}

	// Additional date format validation
	if len(req.StartDate) != 10 || req.StartDate[4] != '-' || req.StartDate[7] != '-' {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid start_date format. Use YYYY-MM-DD format (e.g., 2024-01-01)"))
	}
	if len(req.EndDate) != 10 || req.EndDate[4] != '-' || req.EndDate[7] != '-' {
		return c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid end_date format. Use YYYY-MM-DD format (e.g., 2024-12-31)"))
	}

	response, err := h.attendanceService.GetAttendanceSummary(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Internal server error"))
	}

	return c.JSON(http.StatusOK, dto.NewSuccessResponse("Attendance summary retrieved successfully", response))
}
