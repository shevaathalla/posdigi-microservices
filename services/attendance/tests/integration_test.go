package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"posdigi-attendance/config"
	"posdigi-attendance/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupTestRouter creates a test router for attendance service
func SetupTestRouter() *echo.Echo {
	logger := config.GetLogger()
	e := echo.New()
	e.Validator = middleware.NewCustomValidator()
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(logger))

	// Add health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "attendance-service",
		})
	})

	return e
}

// TestRoute_HealthCheck tests the health check endpoint
func TestRoute_HealthCheck(t *testing.T) {
	e := SetupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "attendance-service", response["service"])
}

// TestRoute_ClockIn_Validation tests clock-in validation
func TestRoute_ClockIn_Validation(t *testing.T) {
	e := SetupTestRouter()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing user_id",
			requestBody: map[string]interface{}{
				"note": "Test clock-in",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Valid request",
			requestBody: map[string]interface{}{
				"user_id": "user-123",
				"note":    "Test clock-in",
			},
			expectedStatus: http.StatusConflict, // User doesn't exist, but validation passes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/attendance/clock-in", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRoute_ClockOut_Validation tests clock-out validation
func TestRoute_ClockOut_Validation(t *testing.T) {
	e := SetupTestRouter()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing attendance_id",
			requestBody: map[string]interface{}{
				"note": "Test clock-out",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Valid request format",
			requestBody: map[string]interface{}{
				"attendance_id": "attend-123",
				"note":          "Test clock-out",
			},
			expectedStatus: http.StatusNotFound, // Attendance doesn't exist, but validation passes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/attendance/clock-out", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRoute_AttendanceHistory_Validation tests attendance history validation
func TestRoute_AttendanceHistory_Validation(t *testing.T) {
	e := SetupTestRouter()

	tests := []struct {
		name           string
		userID         string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Valid request",
			userID:         "user-123",
			queryParams:    "",
			expectedStatus: http.StatusOK, // Empty history is valid
		},
		{
			name:           "With pagination",
			userID:         "user-123",
			queryParams:    "?page=1&limit=5",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid page parameter",
			userID:         "user-123",
			queryParams:    "?page=abc",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Negative page",
			userID:         "user-123",
			queryParams:    "?page=-1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Limit too high",
			userID:         "user-123",
			queryParams:    "?limit=101",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/history/"+tt.userID+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRoute_AttendanceSummary_Validation tests attendance summary validation
func TestRoute_AttendanceSummary_Validation(t *testing.T) {
	e := SetupTestRouter()

	tests := []struct {
		name           string
		userID         string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "Missing start_date",
			userID:         "user-123",
			queryParams:    "?end_date=2024-12-31",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing end_date",
			userID:         "user-123",
			queryParams:    "?start_date=2024-01-01",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing both dates",
			userID:         "user-123",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid date format for start_date",
			userID:         "user-123",
			queryParams:    "?start_date=invalid&end_date=2024-12-31",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid date format for end_date",
			userID:         "user-123",
			queryParams:    "?start_date=2024-01-01&end_date=invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Valid date range",
			userID:         "user-123",
			queryParams:    "?start_date=2024-01-01&end_date=2024-12-31",
			expectedStatus: http.StatusOK, // Empty summary is valid
		},
		{
			name:           "Date format with dashes",
			userID:         "user-123",
			queryParams:    "?start_date=2024-01-01&end_date=2024-12-31",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Different valid date range",
			userID:         "user-123",
			queryParams:    "?start_date=2023-06-15&end_date=2023-06-20",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/attendance/summary/"+tt.userID+tt.queryParams, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestDTO_DateValidation tests various date formats
func TestDTO_DateValidation(t *testing.T) {
	e := echo.New()
	e.Validator = middleware.NewCustomValidator()

	type DateRequest struct {
		StartDate string `json:"start_date" validate:"required"`
		EndDate   string `json:"end_date" validate:"required"`
	}

	e.POST("/test", func(c echo.Context) error {
		var req DateRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, req)
	})

	validDates := []string{
		"2024-01-01",
		"2023-12-31",
		"2020-02-29", // leap year
	}

	for _, date := range validDates {
		t.Run("Valid date: "+date, func(t *testing.T) {
			reqBody := DateRequest{
				StartDate: date,
				EndDate:   date,
			}
			bodyJSON, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}

	invalidDates := []string{
		"2024-13-01", // invalid month
		"2024-01-32", // invalid day
		"2024-01-01T00:00:00", // includes time
		"01-01-2024", // wrong format
		"2024/01/01", // wrong separator
	}

	for _, date := range invalidDates {
		t.Run("Invalid date: "+date, func(t *testing.T) {
			reqBody := DateRequest{
				StartDate: date,
				EndDate:   date,
			}
			bodyJSON, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

// TestMiddleware_ServiceAuth tests that service authentication works
func TestMiddleware_ServiceAuth(t *testing.T) {
	cfg := config.LoadConfig()
	logger := config.GetLogger()
	e := echo.New()
	e.Use(middleware.InternalServiceAuth(cfg.InternalServiceKey, logger))

	e.GET("/protected", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "success"})
	})

	tests := []struct {
		name           string
		serviceKey     string
		expectedStatus int
	}{
		{
			name:           "Valid service key",
			serviceKey:     cfg.InternalServiceKey,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid service key",
			serviceKey:     "wrong-key",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Missing service key",
			serviceKey:     "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.serviceKey != "" {
				req.Header.Set("X-Service-Auth", tt.serviceKey)
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestAttendanceWorkflow tests the complete attendance workflow
func TestAttendanceWorkflow(t *testing.T) {
	e := SetupTestRouter()
	userID := "test-user-" + time.Now().Format("20060102150405")

	// Step 1: Try to clock in (will fail but tests validation)
	clockInReq := map[string]interface{}{
		"user_id": userID,
		"note":    "Starting shift",
	}
	bodyJSON, _ := json.Marshal(clockInReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attendance/clock-in", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	// Should not be a validation error (404 or 409 expected due to missing user)
	assert.NotEqual(t, http.StatusBadRequest, rec.Code, "Clock-in request should pass validation")

	// Step 2: Try to get attendance history (should return empty history)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/attendance/history/"+userID, nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var historyResponse map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &historyResponse)
	require.NoError(t, err)
	assert.True(t, historyResponse["success"].(bool))

	// Step 3: Try to get attendance summary (tests query parameter validation)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/attendance/summary/"+userID+"?start_date=2024-01-01&end_date=2024-12-31", nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var summaryResponse map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &summaryResponse)
	require.NoError(t, err)
	assert.True(t, summaryResponse["success"].(bool))
}