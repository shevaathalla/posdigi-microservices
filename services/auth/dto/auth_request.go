package dto

import (
	"errors"
	"strings"
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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