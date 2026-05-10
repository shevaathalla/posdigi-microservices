package dto

import (
	"errors"
	"strings"
)

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role,omitempty"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	FullName string `json:"full_name,omitempty"`
	Email    string `json:"email,omitempty"`
	Role     string `json:"role,omitempty"`
}

// ListUsersRequest represents a request to list users
type ListUsersRequest struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Search string `json:"search,omitempty"`
}

// Validate validates the create user request
func (r *CreateUserRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("Email is required")
	}
	if strings.TrimSpace(r.FullName) == "" {
		return errors.New("Full name is required")
	}
	if r.Role == "" {
		r.Role = "user" // Default role
	}
	return nil
}

// Validate validates the update user request
func (r *UpdateUserRequest) Validate() error {
	if strings.TrimSpace(r.FullName) == "" && strings.TrimSpace(r.Email) == "" && strings.TrimSpace(r.Role) == "" {
		return errors.New("At least one field must be provided for update")
	}
	return nil
}

// Validate validates the list users request
func (r *ListUsersRequest) Validate() error {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Limit <= 0 {
		r.Limit = 10
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
	return nil
}