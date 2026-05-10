package client

import (
	"context"
	"fmt"

	"posdigi-auth/dto"
)

// UserClient handles all communication with the User Service
type UserClient struct {
	httpClient *HTTPClient
}

// NewUserClient creates a new user service client
func NewUserClient(baseURL string) *UserClient {
	return &UserClient{
		httpClient: NewHTTPClient(baseURL),
	}
}

// CreateUser creates a new user profile in User Service
func (c *UserClient) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	requestBody := map[string]interface{}{
		"email":     req.Email,
		"full_name": req.FullName,
		"role":      req.Role,
	}

	var response dto.CreateUserResponse
	if err := c.httpClient.post(ctx, "/api/v1/users", requestBody, &response); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &response, nil
}

// GetUserByID retrieves a user profile by ID
func (c *UserClient) GetUserByID(ctx context.Context, userID string) (*dto.UserProfileResponse, error) {
	var response dto.UserProfileResponse
	if err := c.httpClient.get(ctx, fmt.Sprintf("/api/v1/users/%s", userID), &response); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &response, nil
}