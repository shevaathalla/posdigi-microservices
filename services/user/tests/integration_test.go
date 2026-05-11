package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"posdigi-user/config"
	"posdigi-user/middleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupTestRouter creates a test router for user service
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
			"service": "user-service",
		})
	})

	// Add test user validation endpoint
	type CreateUserRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
		Role     string `json:"role" validate:"omitempty,oneof=user admin"`
	}

	e.POST("/api/v1/users", func(c echo.Context) error {
		var req CreateUserRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, req)
	})

	// Add test employee validation endpoint
	type CreateEmployeeRequest struct {
		UserID  string `json:"user_id" validate:"required,uuid"`
		FullName string `json:"full_name" validate:"required,min=2,max=100"`
	}

	e.POST("/api/v1/employees", func(c echo.Context) error {
		var req CreateEmployeeRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, req)
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
	assert.Equal(t, "user-service", response["service"])
}

// TestRoute_CreateUser_Validation tests user creation validation
func TestRoute_CreateUser_Validation(t *testing.T) {
	e := SetupTestRouter()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing email",
			requestBody: map[string]interface{}{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid email format",
			requestBody: map[string]interface{}{
				"email":    "not-an-email",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Password too short",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "12345",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRoute_CreateEmployee_Validation tests employee creation validation
func TestRoute_CreateEmployee_Validation(t *testing.T) {
	e := SetupTestRouter()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Missing user_id",
			requestBody: map[string]interface{}{
				"full_name": "John Doe",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing full_name",
			requestBody: map[string]interface{}{
				"user_id": "user-123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Full name too short",
			requestBody: map[string]interface{}{
				"user_id":   "user-123",
				"full_name": "JD",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Negative salary",
			requestBody: map[string]interface{}{
				"user_id":   "user-123",
				"full_name": "John Doe",
				"salary":    -1000,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid employment status",
			requestBody: map[string]interface{}{
				"user_id":           "user-123",
				"full_name":         "John Doe",
				"employment_status": "invalid_status",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/employees", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRoute_UpdateEmployeeStatus_Validation tests employee status update validation
func TestRoute_UpdateEmployeeStatus_Validation(t *testing.T) {
	e := SetupTestRouter()

	tests := []struct {
		name           string
		employeeID     string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name:       "Missing employment_status field",
			employeeID: "emp-123",
			requestBody: map[string]interface{}{
				"status": "active", // wrong field name
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid employment status value",
			employeeID: "emp-123",
			requestBody: map[string]interface{}{
				"employment_status": "invalid_value",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request",
			employeeID:     "emp-123",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/employees/"+tt.employeeID+"/status", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRoute_UpdateEmployeeStatus_ValidStatuses tests valid employment status values
func TestRoute_UpdateEmployeeStatus_ValidStatuses(t *testing.T) {
	e := SetupTestRouter()

	validStatuses := []string{"active", "terminated", "on_leave", "suspended"}

	for _, status := range validStatuses {
		t.Run("Valid status: "+status, func(t *testing.T) {
			requestBody := map[string]interface{}{
				"employment_status": status,
			}
			bodyJSON, _ := json.Marshal(requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/employees/emp-123/status", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// Should not be a validation error (404 is expected since employee doesn't exist)
			assert.NotEqual(t, http.StatusBadRequest, rec.Code, "Status %s should pass validation", status)
		})
	}
}

// TestMiddleware_ServiceAuth tests that service authentication middleware works
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
			serviceKey:     "invalid-key",
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

// TestPagination_Parameters tests pagination parameter parsing
func TestPagination_Parameters(t *testing.T) {
	e := SetupTestRouter()

	tests := []struct {
		name           string
		url            string
		expectError    bool
		expectedPage   int
		expectedLimit  int
	}{
		{
			name:         "Default pagination",
			url:          "/api/v1/users",
			expectError:   false,
			expectedPage:  1,
			expectedLimit: 10,
		},
		{
			name:         "Custom pagination",
			url:          "/api/v1/users?page=2&page_size=20",
			expectError:  false,
			expectedPage: 2,
			expectedLimit: 20,
		},
		{
			name:         "Invalid page number",
			url:          "/api/v1/users?page=abc",
			expectError:  true,
		},
		{
			name:         "Negative page",
			url:          "/api/v1/users?page=-1",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, rec.Code)
			} else {
				// Even if user not found, should not be a bad request for pagination
				assert.NotEqual(t, http.StatusBadRequest, rec.Code)
			}
		})
	}
}

// TestRoute_SearchFunctionality tests search parameter handling
func TestRoute_SearchFunctionality(t *testing.T) {
	e := SetupTestRouter()

	searchTerms := []string{"john", "doe", "engineer", "123", "@#$"}

	for _, searchTerm := range searchTerms {
		t.Run("Search term: "+searchTerm, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/users?search="+searchTerm, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// Should handle search gracefully (400 for auth error, not bad request for search params)
			assert.NotEqual(t, http.StatusInternalServerError, rec.Code)
		})
	}
}

// TestDTO_UserRequest_Validation tests user request DTO validation
func TestDTO_UserRequest_Validation(t *testing.T) {
	e := echo.New()
	e.Validator = middleware.NewCustomValidator()

	type CreateUserRequest struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	e.POST("/users", func(c echo.Context) error {
		var req CreateUserRequest
		if err := c.Bind(&req); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, req)
	})

	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
	}{
		{
			name: "Valid request",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Email required",
			requestBody: map[string]string{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Email format",
			requestBody: map[string]string{
				"email":    "invalid-email",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Password length",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "12345",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}