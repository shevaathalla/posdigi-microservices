package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"posdigi-auth/config"
)

// HTTPClient handles HTTP communication between microservices
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	serviceKey string
}

// NewHTTPClient creates a new HTTP client for service communication
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL:    baseURL,
		serviceKey: config.GetEnv("INTERNAL_SERVICE_KEY", "internal-service-key"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// post sends a POST request expecting 201 Created (used for resource creation)
func (c *HTTPClient) post(ctx context.Context, path string, body interface{}, response interface{}) error {
	return c.postWithStatus(ctx, path, body, http.StatusCreated, response)
}

// postOK sends a POST request expecting 200 OK (used for actions like authenticate)
func (c *HTTPClient) postOK(ctx context.Context, path string, body interface{}, response interface{}) error {
	return c.postWithStatus(ctx, path, body, http.StatusOK, response)
}

// postWithStatus sends a POST request with a configurable expected status
func (c *HTTPClient) postWithStatus(ctx context.Context, path string, body interface{}, expectedStatus int, response interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Auth", c.serviceKey)

	return c.do(req, expectedStatus, response)
}

// get sends a GET request to the user service
func (c *HTTPClient) get(ctx context.Context, path string, response interface{}) error {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Auth", c.serviceKey)

	return c.do(req, http.StatusOK, response)
}

// delete sends a DELETE request to the user service
func (c *HTTPClient) delete(ctx context.Context, path string, response interface{}) error {
	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Auth", c.serviceKey)

	return c.do(req, http.StatusOK, response)
}

// serviceError holds a parsed error message from a downstream service response
type serviceError struct {
	Message string `json:"message"`
}

// do executes the HTTP request and handles the response
func (c *HTTPClient) do(req *http.Request, expectedStatus int, response interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Don't expose internal network details — log them, return a clean error
		config.Errorf("HTTP request to %s failed: %v", req.URL.Path, err)
		return fmt.Errorf("service unavailable")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body")
	}

	// Non-2xx / unexpected status: extract message from downstream JSON if possible
	if resp.StatusCode != expectedStatus {
		var errResp serviceError
		if jsonErr := json.Unmarshal(body, &errResp); jsonErr == nil && errResp.Message != "" {
			return fmt.Errorf("%s", errResp.Message)
		}
		// Fallback: just report the status code without leaking raw body
		config.Errorf("Unexpected status %d from %s: %s", resp.StatusCode, req.URL.Path, string(body))
		return fmt.Errorf("upstream service returned status %d", resp.StatusCode)
	}

	// Parse the standard response wrapper { "success": bool, "data": {...} }
	var wrapper struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		config.Errorf("Failed to decode wrapper from %s: %v", req.URL.Path, err)
		return fmt.Errorf("invalid response from upstream service")
	}

	if !wrapper.Success {
		if wrapper.Message != "" {
			return fmt.Errorf("%s", wrapper.Message)
		}
		return fmt.Errorf("upstream service returned an unsuccessful response")
	}

	// Decode the inner data directly into the target struct
	if response != nil && len(wrapper.Data) > 0 {
		if err := json.Unmarshal(wrapper.Data, response); err != nil {
			config.Errorf("Failed to decode data from %s: %v", req.URL.Path, err)
			return fmt.Errorf("failed to decode upstream response data")
		}
	}

	return nil
}
