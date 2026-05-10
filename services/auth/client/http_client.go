package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"posdigi-auth/dto"
)

// HTTPClient handles HTTP communication between microservices
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHTTPClient creates a new HTTP client for service communication
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetUserProfile fetches user profile from User Service
func (c *HTTPClient) GetUserProfile(ctx context.Context, userID string) (*dto.UserProfileResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Auth", "internal-service-key") // Internal service auth

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var userProfile struct {
		Success bool                     `json:"success"`
		Data    dto.UserProfileResponse  `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userProfile); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !userProfile.Success {
		return nil, fmt.Errorf("user service returned unsuccessful response")
	}

	return &userProfile.Data, nil
}

// CreateUserProfile creates a new user profile in User Service
func (c *HTTPClient) CreateUserProfile(ctx context.Context, email, fullName string) (*dto.CreateUserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users", c.baseURL)

	requestBody := map[string]interface{}{
		"email":     email,
		"full_name": fullName,
		"role":      "user",
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Auth", "internal-service-key")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("user service returned status %d", resp.StatusCode)
	}

	var createUserResp struct {
		Success bool                  `json:"success"`
		Data    dto.CreateUserResponse `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&createUserResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !createUserResp.Success {
		return nil, fmt.Errorf("user service returned unsuccessful response")
	}

	return &createUserResp.Data, nil
}