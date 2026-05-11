package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"posdigi-auth/config"
	"posdigi-auth/dto"
	"posdigi-auth/middleware"
)

// TestRoute_HealthCheck tests the health check endpoint
func TestRoute_HealthCheck(t *testing.T) {
	logger := config.GetLogger()
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(logger))

	// Add health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "healthy",
			"service": "auth-service",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "auth-service", response["service"])
}

// TestRoute_Register_Validation tests registration validation
func TestRoute_Register_Validation(t *testing.T) {
	logger := config.GetLogger()
	e := echo.New()
	e.Validator = middleware.NewCustomValidator()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger(logger))

	// Create a simple test endpoint that validates the request
	e.POST("/api/v1/auth/register", func(c echo.Context) error {
		var req dto.RegisterRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, req)
	})

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestRoute_Login_Validation tests login validation
func TestRoute_Login_Validation(t *testing.T) {
	logger := config.GetLogger()
	e := echo.New()
	e.Validator = middleware.NewCustomValidator()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger(logger))

	// Create a simple test endpoint that validates the request
	e.POST("/api/v1/auth/login", func(c echo.Context) error {
		var req dto.LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		}
		if err := c.Validate(&req); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, req)
	})

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(bodyJSON))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// TestDTO_Validation tests DTO validation rules
func TestDTO_Validation(t *testing.T) {
	validate := validator.New()

	t.Run("RegisterRequest validation", func(t *testing.T) {
		// Test valid request
		validReq := dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		err := validate.Struct(validReq)
		assert.NoError(t, err)

		// Test invalid email
		invalidReq := dto.RegisterRequest{
			Email:    "not-an-email",
			Password: "password123",
		}

		err = validate.Struct(invalidReq)
		assert.Error(t, err)
	})

	t.Run("LoginRequest validation", func(t *testing.T) {
		// Test valid request
		validReq := dto.LoginRequest{
			Email:    "test@example.com",
			Password: "anypassword",
		}

		err := validate.Struct(validReq)
		assert.NoError(t, err)

		// Test missing password
		invalidReq := dto.LoginRequest{
			Email: "test@example.com",
		}

		err = validate.Struct(invalidReq)
		assert.Error(t, err)
	})

	t.Run("TokenRequest validation", func(t *testing.T) {
		// Test valid request
		validReq := dto.TokenRequest{
			Token: "valid-token",
		}

		err := validate.Struct(validReq)
		assert.NoError(t, err)

		// Test missing token
		invalidReq := dto.TokenRequest{}

		err = validate.Struct(invalidReq)
		assert.Error(t, err)
	})
}

// TestMiddleware_CustomValidator tests custom validator functionality
func TestMiddleware_CustomValidator(t *testing.T) {
	e := echo.New()
	e.Validator = middleware.NewCustomValidator()

	type TestStruct struct {
		Email string `json:"email" validate:"required,email"`
		Name  string `json:"name" validate:"required,min=2"`
	}

	e.POST("/test", func(c echo.Context) error {
		var req TestStruct
		if err := c.Bind(&req); err != nil {
			return err
		}
		if err := c.Validate(&req); err != nil {
			return err
		}
		return c.JSON(http.StatusOK, req)
	})

	// Test successful binding
	t.Run("Valid request", func(t *testing.T) {
		req := map[string]string{
			"email": "test@example.com",
			"name":  "John Doe",
		}
		bodyJSON, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(bodyJSON))
		httpReq.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, httpReq)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	// Test validation error
	t.Run("Invalid email", func(t *testing.T) {
		req := map[string]string{
			"email": "not-an-email",
			"name":  "John Doe",
		}
		bodyJSON, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(bodyJSON))
		httpReq.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, httpReq)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	// Test name too short
	t.Run("Name too short", func(t *testing.T) {
		req := map[string]string{
			"email": "test@example.com",
			"name":  "J",
		}
		bodyJSON, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(bodyJSON))
		httpReq.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, httpReq)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

// TestMiddleware_RequestID tests request ID generation
func TestMiddleware_RequestID(t *testing.T) {
	logger := config.GetLogger()
	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(logger))

	e.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "test",
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

// TestMiddleware_Recover tests panic recovery
func TestMiddleware_Recover(t *testing.T) {
	e := echo.New()
	e.Use(middleware.Recover())

	e.GET("/panic", func(c echo.Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should recover and return 500 instead of crashing
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestEmailValidationFormats tests various email formats
func TestEmailValidationFormats(t *testing.T) {
	validate := validator.New()

	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.com",
		"user123@test-site.com",
		"a@b.c",
	}

	for _, email := range validEmails {
		t.Run("Valid email: "+email, func(t *testing.T) {
		type EmailTest struct {
			Email string `validate:"required,email"`
		}
		req := EmailTest{Email: email}
			err := validate.Struct(req)
			assert.NoError(t, err, "Email should be valid: "+email)
		})
	}

	invalidEmails := []string{
		"plainaddress",
		"@missinglocal.com",
		"username@",
		"username@.com",
		"username@com",
	}

	for _, email := range invalidEmails {
		t.Run("Invalid email: "+email, func(t *testing.T) {
			type EmailTest struct {
				Email string `validate:"required,email"`
			}
			req := EmailTest{Email: email}
			err := validate.Struct(req)
			assert.Error(t, err, "Email should be invalid: "+email)
		})
	}
}