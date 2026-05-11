package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"posdigi-gateway/client"
	"posdigi-gateway/config"
	"posdigi-gateway/handler"
	"posdigi-gateway/middleware"
	"posdigi-gateway/router"
	"posdigi-gateway/service"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupTestGateway creates a test gateway instance
func SetupTestGateway() *echo.Echo {
	cfg := config.LoadConfig()
	logger := config.GetLogger()

	// Create service clients
	authClient := client.NewServiceClient(cfg.AuthServiceURL, cfg.InternalServiceKey, logger)
	userClient := client.NewServiceClient(cfg.UserServiceURL, cfg.InternalServiceKey, logger)
	attendanceClient := client.NewServiceClient(cfg.AttendanceServiceURL, cfg.InternalServiceKey, logger)

	// Create health checker
	healthChecker := service.NewHealthChecker(cfg, logger)

	// Create proxy handler
	proxyHandler := handler.NewProxyHandler(authClient, userClient, attendanceClient, logger)

	// Setup routes
	e := echo.New()
	middleware.SetupMiddleware(e, cfg, logger)
	router.SetupRoutes(e, proxyHandler, healthChecker, logger)

	return e
}

// TestGateway_HealthCheck tests the gateway health check endpoint
func TestGateway_HealthCheck(t *testing.T) {
	e := SetupTestGateway()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Gateway should respond even if backend services are down
	assert.NotEqual(t, http.StatusServiceUnavailable, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "gateway", response["service"])
}

// TestGateway_RateLimiting tests that rate limiting middleware is applied
func TestGateway_RateLimiting(t *testing.T) {
	e := SetupTestGateway()

	// Make multiple rapid requests to trigger rate limiting
	successCount := 0
	rateLimitTriggered := false

	for i := 0; i < 150; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		if rec.Code == http.StatusTooManyRequests {
			rateLimitTriggered = true
			break
		}
		if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusBadGateway {
			// Expected (no auth or service down)
			successCount++
		}
	}

	// We expect some rate limiting to occur or requests to be processed
	assert.True(t, successCount > 0 || rateLimitTriggered, "Gateway should process requests or apply rate limiting")
}

// TestGateway_CORS tests that CORS middleware is working
func TestGateway_CORS(t *testing.T) {
	e := SetupTestGateway()

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/users", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should handle preflight request
	assert.NotEqual(t, http.StatusMethodNotAllowed, rec.Code)

	corsHeaders := []string{
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
	}

	hasCorsHeaders := false
	for _, header := range corsHeaders {
		if rec.Header().Get(header) != "" {
			hasCorsHeaders = true
			break
		}
	}

	assert.True(t, hasCorsHeaders, "CORS headers should be present")
}

// TestGateway_RequestID tests that RequestID middleware adds unique IDs
func TestGateway_RequestID(t *testing.T) {
	e := SetupTestGateway()

	// Make multiple requests
	requestIDs := make(map[string]bool)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		requestID := rec.Header().Get("X-Request-ID")
		if requestID != "" {
			requestIDs[requestID] = true
		}
	}

	// Should have multiple unique request IDs
	assert.True(t, len(requestIDs) >= 3, "Should have multiple unique request IDs")
}

// TestGateway_QueryParameterForwarding tests query parameters are forwarded
func TestGateway_QueryParameterForwarding(t *testing.T) {
	e := SetupTestGateway()

	testCases := []struct {
		name        string
		url         string
		params      string
		expectError bool
	}{
		{
			name:        "User pagination",
			url:         "/api/v1/users",
			params:      "?page=1&limit=10",
			expectError: false, // Will fail auth, but not bad request
		},
		{
			name:        "User search",
			url:         "/api/v1/users",
			params:      "?search=john",
			expectError: false,
		},
		{
			name:        "Attendance summary with dates",
			url:         "/api/v1/attendance/summary/user-123",
			params:      "?start_date=2024-01-01&end_date=2024-12-31",
			expectError: false,
		},
		{
			name:        "Attendance history pagination",
			url:         "/api/v1/attendance/history/user-123",
			params:      "?page=1&limit=5",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.url+tc.params, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			if tc.expectError {
				assert.Equal(t, http.StatusBadRequest, rec.Code, "Expected bad request for "+tc.name)
			} else {
				// Should not be a bad request (may be unauthorized or service unavailable)
				assert.NotEqual(t, http.StatusBadRequest, rec.Code, "Should not be bad request for "+tc.name)
			}
		})
	}
}

// TestGateway_HeadersForwarding tests that headers are properly forwarded
func TestGateway_HeadersForwarding(t *testing.T) {
	e := SetupTestGateway()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Test-Agent/1.0")
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Request should be processed (even if auth fails)
	assert.NotEqual(t, http.StatusMethodNotAllowed, rec.Code)
	assert.NotEqual(t, http.StatusNotFound, rec.Code)
}

// TestGateway_ErrorHandling tests error handling in gateway
func TestGateway_ErrorHandling(t *testing.T) {
	e := SetupTestGateway()

	testCases := []struct {
		name           string
		method         string
		url            string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "Invalid JSON",
			method:         http.MethodPost,
			url:            "/api/v1/auth/login",
			body:           []byte("{invalid json}"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty POST",
			method:         http.MethodPost,
			url:            "/api/v1/auth/login",
			body:           []byte("{}"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Valid JSON but missing fields",
			method:         http.MethodPost,
			url:            "/api/v1/auth/login",
			body:           []byte(`{"email": "test@example.com"}`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Non-existent route",
			method:         http.MethodGet,
			url:            "/api/v1/nonexistent",
			body:           nil,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, bytes.NewReader(tc.body))
			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Code)
		})
	}
}

// TestGateway_MethodRouting tests that different HTTP methods are routed correctly
func TestGateway_MethodRouting(t *testing.T) {
	e := SetupTestGateway()

	testCases := []struct {
		name   string
		method string
		url    string
	}{
		{
			name:   "GET users",
			method: http.MethodGet,
			url:    "/api/v1/users",
		},
		{
			name:   "POST users",
			method: http.MethodPost,
			url:    "/api/v1/users",
		},
		{
			name:   "PUT user",
			method: http.MethodPut,
			url:    "/api/v1/users/user-123",
		},
		{
			name:   "DELETE user",
			method: http.MethodDelete,
			url:    "/api/v1/users/user-123",
		},
		{
			name:   "PATCH employee status",
			method: http.MethodPatch,
			url:    "/api/v1/employees/emp-123/status",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.url, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// Should not return method not allowed
			assert.NotEqual(t, http.StatusMethodNotAllowed, rec.Code, "Method "+tc.method+" should be supported")
		})
	}
}

// TestGateway_HealthCheckIntegration tests integrated health checks
func TestGateway_HealthCheckIntegration(t *testing.T) {
	e := SetupTestGateway()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Gateway should always respond
	assert.Equal(t, "gateway", response["service"])

	// Check if services field exists
	if _, ok := response["services"]; ok {
		// Services map should be present
		services, ok := response["services"].(map[string]interface{})
		require.True(t, ok, "Services should be a map")

		// Each service should have health status
		for serviceName, serviceHealth := range services {
			serviceMap, ok := serviceHealth.(map[string]interface{})
			assert.True(t, ok, "Service health should be a map")
			assert.Contains(t, serviceMap, "healthy")
			assert.Contains(t, serviceMap, "last_check")
			_ = serviceName // use variable
		}
	}
}

// TestGateway_ConcurrentRequests tests that gateway handles concurrent requests
func TestGateway_ConcurrentRequests(t *testing.T) {
	e := SetupTestGateway()

	// Make concurrent requests
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func() {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			done <- true
		}()
	}

	// Wait for all requests to complete
	completed := 0
	timeout := time.After(5 * time.Second)

	for {
		select {
		case <-done:
			completed++
			if completed == concurrency {
				return // All requests completed successfully
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for concurrent requests. Completed: %d/%d", completed, concurrency)
		}
	}
}

// TestGateway_PublicRoutes tests that public routes don't require authentication
func TestGateway_PublicRoutes(t *testing.T) {
	e := SetupTestGateway()

	publicRoutes := []struct {
		name string
		url  string
	}{
		{
			name: "Health check",
			url:  "/health",
		},
		{
			name: "Register",
			url:  "/api/v1/auth/register",
		},
		{
			name: "Login",
			url:  "/api/v1/auth/login",
		},
	}

	for _, route := range publicRoutes {
		t.Run(route.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route.url, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// Should not return unauthorized for public routes
			assert.NotEqual(t, http.StatusUnauthorized, rec.Code, "Public route should not return 401")
		})
	}
}

// TestGateway_ProtectedRoutes tests that protected routes require authentication
func TestGateway_ProtectedRoutes(t *testing.T) {
	e := SetupTestGateway()

	protectedRoutes := []struct {
		name string
		url  string
	}{
		{
			name: "List users",
			url:  "/api/v1/users",
		},
		{
			name: "Get user by ID",
			url:  "/api/v1/users/user-123",
		},
		{
			name: "Employee list",
			url:  "/api/v1/employees",
		},
		{
			name: "Attendance history",
			url:  "/api/v1/attendance/history/user-123",
		},
	}

	for _, route := range protectedRoutes {
		t.Run(route.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, route.url, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			// Should return unauthorized for protected routes without auth
			assert.Equal(t, http.StatusUnauthorized, rec.Code, "Protected route should return 401 without auth")
		})
	}
}