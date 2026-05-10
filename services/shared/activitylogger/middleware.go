package activitylogger

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// Middleware provides automatic activity logging for HTTP requests
type Middleware struct {
	logger *Logger
}

// NewMiddleware creates a new activity logging middleware
func NewMiddleware(logger *Logger) *Middleware {
	return &Middleware{
		logger: logger,
	}
}

// ActivityLoggingMiddleware returns Echo middleware for automatic activity logging
func (m *Middleware) ActivityLoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract request information
			requestID := c.Response().Header().Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = c.Response().Header().Get(echo.HeaderXRequestID)
			}
			if requestID == "" {
				requestID = c.Request().Header.Get("X-Request-ID")
			}

			// Skip logging for health checks and swagger docs
			if m.shouldSkipLogging(c.Request().URL.Path) {
				return next(c)
			}

			// Extract user information if available
			userID := m.extractUserID(c)

			// Capture request details
			startTime := time.Now()
			method := c.Request().Method
			path := c.Request().URL.Path
			ip := c.RealIP()
			userAgent := c.Request().UserAgent()

			// Read and restore request body if needed
			var bodyBytes []byte
			if c.Request().Body != nil && c.Request().Body != http.NoBody {
				bodyBytes, _ = io.ReadAll(c.Request().Body)
				c.Request().Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}

			// Create custom response writer to capture status code
			wrapper := &responseWrapper{ResponseWriter: c.Response().Writer}
			c.Response().Writer = wrapper

			// Continue with the request
			err := next(c)

			// Restore original writer
			c.Response().Writer = wrapper.ResponseWriter

			// Calculate duration
			duration := time.Since(startTime)
			statusCode := c.Response().Status

			// Log the activity asynchronously
			go m.logActivity(context.Background(), ActivityLog{
				RequestID:  requestID,
				UserID:     userID,
				Service:    m.logger.service,
				Action:     m.determineAction(method, path),
				Method:     method,
				Endpoint:   path,
				IPAddress:  ip,
				UserAgent:  userAgent,
				StatusCode: statusCode,
				Success:    statusCode >= 200 && statusCode < 400,
				Timestamp:  startTime,
				CreatedAt:  time.Now(),
			}, err, duration)

			return err
		}
	}
}

// shouldSkipLogging determines if a request should be skipped from logging
func (m *Middleware) shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/health",
		"/docs",
		"/swagger",
		"/favicon.ico",
		"/metrics",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// extractUserID attempts to extract user ID from context
func (m *Middleware) extractUserID(c echo.Context) string {
	// Try to get user ID from various context sources
	if userID := c.Param("id"); userID != "" {
		return userID
	}

	if userID := c.QueryParam("user_id"); userID != "" {
		return userID
	}

	// Try to get from context (set by auth middleware)
	if userID := c.Get("user_id"); userID != nil {
		if str, ok := userID.(string); ok {
			return str
		}
	}

	if userID := c.Get("userId"); userID != nil {
		if str, ok := userID.(string); ok {
			return str
		}
	}

	return ""
}

// determineAction determines the action type based on method and path
func (m *Middleware) determineAction(method, path string) string {
	// Extract action from path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) > 0 {
		resource := pathParts[len(pathParts)-1]
		if strings.Contains(resource, "-") {
			parts := strings.Split(resource, "-")
			resource = parts[len(parts)-1]
		}

		switch method {
		case "GET":
			if strings.Contains(path, "login") || strings.Contains(path, "authenticate") {
				return ActionLoginSuccess
			}
			return "VIEW_" + strings.ToUpper(resource)
		case "POST":
			if strings.Contains(path, "login") {
				return ActionLoginSuccess
			}
			if strings.Contains(path, "logout") {
				return ActionLogout
			}
			if strings.Contains(path, "register") {
				return ActionRegister
			}
			if strings.Contains(path, "clock-in") {
				return ActionClockIn
			}
			if strings.Contains(path, "clock-out") {
				return ActionClockOut
			}
			return "CREATE_" + strings.ToUpper(resource)
		case "PUT", "PATCH":
			return "UPDATE_" + strings.ToUpper(resource)
		case "DELETE":
			return "DELETE_" + strings.ToUpper(resource)
		}
	}

	return method + "_" + path
}

// logActivity logs activity asynchronously
func (m *Middleware) logActivity(ctx context.Context, log ActivityLog, err error, duration time.Duration) {
	// Add error information if request failed
	if err != nil && log.Success {
		log.Success = false
		log.ErrorMessage = err.Error()
	}

	// Add metadata
	if log.Metadata == nil {
		log.Metadata = &ActivityMetadata{}
	}

	log.Metadata.Extra = map[string]interface{}{
		"duration_ms": duration.Milliseconds(),
	}

	// Set ID if not set
	if log.ID.IsZero() {
		log.ID = mongoInsertID()
	}

	// Log asynchronously
	_ = m.logger.LogActivity(ctx, &log)
}

// responseWrapper wraps http.ResponseWriter to capture status code
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}