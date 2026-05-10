package dto

import (
	"errors"
	"strings"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email        string                `json:"email"`
	Password     string                `json:"password"`
	EmployeeData *EmployeeDataRequest  `json:"employee_data,omitempty"`
}

// EmployeeDataRequest represents employee profile data during registration
type EmployeeDataRequest struct {
	FullName         string  `json:"full_name,omitempty"`
	Phone            string  `json:"phone,omitempty"`
	Department       string  `json:"department,omitempty"`
	Position         string  `json:"position,omitempty"`
	Salary           float64 `json:"salary,omitempty"`
	HireDate         string  `json:"hire_date,omitempty"`
	EmploymentStatus string  `json:"employment_status,omitempty"`
	ManagerID        *string `json:"manager_id,omitempty"`
	EmergencyContact string  `json:"emergency_contact,omitempty"`
	EmergencyPhone   string  `json:"emergency_phone,omitempty"`
	Address          string  `json:"address,omitempty"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenRequest represents a token validation request
type TokenRequest struct {
	Token string `json:"token"`
}

// Validate validates the register request
func (r *RegisterRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("Email is required")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("Password is required")
	}
	if len(r.Password) < 6 {
		return errors.New("Password must be at least 6 characters")
	}
	return nil
}

// Validate validates the login request
func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" {
		return errors.New("Email is required")
	}
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("Password is required")
	}
	return nil
}

// Validate validates the token request
func (r *TokenRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return errors.New("Token is required")
	}
	return nil
}