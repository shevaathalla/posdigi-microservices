package dto

// AuthResponse represents a standard API response
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// UserResponse represents user information in responses
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

// TokenValidationResponse represents a token validation response
type TokenValidationResponse struct {
	Valid   bool   `json:"valid"`
	UserID  string `json:"user_id,omitempty"`
	Email   string `json:"email,omitempty"`
	Role    string `json:"role,omitempty"`
	Message string `json:"message,omitempty"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(message string, data any) AuthResponse {
	return AuthResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(message string) AuthResponse {
	return AuthResponse{
		Success: false,
		Message: message,
	}
}

// NewUserResponse creates a user response from repository user
func NewUserResponse(id, email, role string) UserResponse {
	return UserResponse{
		ID:    id,
		Email: email,
		Role:  role,
	}
}