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

// post sends a POST request to the user service
func (c *HTTPClient) post(ctx context.Context, path string, body interface{}, response interface{}) error {
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

	return c.do(req, http.StatusCreated, response)
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

// do executes the HTTP request and handles the response
func (c *HTTPClient) do(req *http.Request, expectedStatus int, response interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response wrapper
	var wrapper struct {
		Success bool `json:"success"`
		Data    any  `json:"data"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		return fmt.Errorf("failed to decode wrapper: %w", err)
	}

	if !wrapper.Success {
		return fmt.Errorf("unsuccessful response")
	}

	// Extract the actual data
	dataWrapper := map[string]interface{}{
		"Data": wrapper.Data,
	}

	dataBytes, err := json.Marshal(dataWrapper)
	if err != nil {
		return fmt.Errorf("failed to remarshal data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
