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
func (c *UserClient) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserProfileResponse, error) {
	requestBody := map[string]interface{}{
		"email":     req.Email,
		"password":  req.Password,
		"full_name": req.FullName,
		"role":      req.Role,
	}

	var response dto.UserProfileResponse
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

// GetUserByEmail retrieves a user profile by email
func (c *UserClient) GetUserByEmail(ctx context.Context, email string) (*dto.UserProfileResponse, error) {
	var response dto.UserProfileResponse
	if err := c.httpClient.get(ctx, fmt.Sprintf("/api/v1/users/email/%s", email), &response); err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &response, nil
}

// AuthenticateUser validates credentials via User Service
func (c *UserClient) AuthenticateUser(ctx context.Context, email, password string) (*dto.UserProfileResponse, error) {
	requestBody := map[string]interface{}{
		"email":    email,
		"password": password,
	}

	var response dto.UserProfileResponse
	if err := c.httpClient.post(ctx, "/api/v1/users/authenticate", requestBody, &response); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	return &response, nil
}

// CreateEmployee creates a new employee profile in User Service
func (c *UserClient) CreateEmployee(ctx context.Context, userID string, employeeData *dto.EmployeeDataRequest) error {
	requestBody := map[string]interface{}{
		"user_id":           userID,
		"full_name":         employeeData.FullName,
		"phone":             employeeData.Phone,
		"department":        employeeData.Department,
		"position":          employeeData.Position,
		"salary":            employeeData.Salary,
		"hire_date":         employeeData.HireDate,
		"employment_status": employeeData.EmploymentStatus,
		"manager_id":        employeeData.ManagerID,
		"emergency_contact": employeeData.EmergencyContact,
		"emergency_phone":   employeeData.EmergencyPhone,
		"address":           employeeData.Address,
	}

	var response map[string]interface{}
	if err := c.httpClient.post(ctx, "/api/v1/employees", requestBody, &response); err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}

	return nil
}

// DeleteUser deletes a user profile (for rollback scenarios)
func (c *UserClient) DeleteUser(ctx context.Context, userID string) error {
	var response interface{}
	if err := c.httpClient.delete(ctx, fmt.Sprintf("/api/v1/users/%s", userID), &response); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}