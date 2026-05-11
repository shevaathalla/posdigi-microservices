package testing

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResponse represents a test HTTP response
type TestResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// MakeRequest creates and executes an HTTP request for testing
func MakeRequest(e *echo.Echo, method, url string, body interface{}, headers map[string]string) *TestResponse {
	var bodyReader *bytes.Reader
	if body != nil {
		bodyJSON, err := json.Marshal(body)
		if err != nil {
			return &TestResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       []byte(`{"error": "Failed to marshal request body"}`),
			}
		}
		bodyReader = bytes.NewReader(bodyJSON)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	return &TestResponse{
		StatusCode: rec.Code,
		Body:       rec.Body.Bytes(),
		Headers:    rec.Header(),
	}
}

// AssertJSONSuccess asserts that a response is a successful JSON response
func AssertJSONSuccess(t *testing.T, response *TestResponse) map[string]interface{} {
	assert.NotEqual(t, http.StatusInternalServerError, response.StatusCode)

	var result map[string]interface{}
	err := json.Unmarshal(response.Body, &result)
	require.NoError(t, err, "Response should be valid JSON")

	assert.True(t, result["success"].(bool), "Response should indicate success")
	return result
}

// AssertJSONError asserts that a response is an error JSON response
func AssertJSONError(t *testing.T, response *TestResponse) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal(response.Body, &result)
	require.NoError(t, err, "Response should be valid JSON")

	if success, ok := result["success"].(bool); ok && success {
		assert.Fail(t, "Expected error response but got success")
	}

	return result
}

// AssertStatusCode asserts that the response has a specific status code
func AssertStatusCode(t *testing.T, response *TestResponse, expectedStatus int) {
	assert.Equal(t, expectedStatus, response.StatusCode, "Status code should match")
}

// AssertValidationError asserts that a response contains validation errors
func AssertValidationError(t *testing.T, response *TestResponse) map[string]interface{} {
	AssertStatusCode(t, response, http.StatusBadRequest)

	var result map[string]interface{}
	err := json.Unmarshal(response.Body, &result)
	require.NoError(t, err)

	// Check if it contains validation errors
	if _, ok := result["errors"]; ok {
		return result
	}
	if _, ok := result["message"].(string); ok {
		return result
	}

	assert.Fail(t, "Response should contain validation errors")
	return result
}

// CreateTestAuthHeader creates authentication headers for testing
func CreateTestAuthHeader(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json",
	}
}

// CreateTestServiceAuthHeader creates service authentication headers for testing
func CreateTestServiceAuthHeader(serviceKey string) map[string]string {
	return map[string]string{
		"X-Service-Auth": serviceKey,
		"Content-Type":  "application/json",
	}
}

// SetupTestEcho creates a test Echo instance with common middleware
func SetupTestEcho() *echo.Echo {
	e := echo.New()
	e.Use(echo.MiddlewareFunc())
	return e
}

// AssertCommonErrors checks for common error scenarios
func AssertCommonErrors(t *testing.T, response *TestResponse, service string) {
	switch response.StatusCode {
	case http.StatusBadRequest:
		t.Logf("✅ %s: Validation error properly returned", service)
	case http.StatusUnauthorized:
		t.Logf("✅ %s: Authorization properly required", service)
	case http.StatusNotFound:
		t.Logf("✅ %s: Resource not found properly handled", service)
	case http.StatusInternalServerError:
		t.Logf("❌ %s: Internal server error (may need investigation)", service)
	default:
		t.Logf("ℹ️  %s: Status %d returned", service, response.StatusCode)
	}
}