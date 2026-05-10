package dto

// UserResponse represents a user profile in responses
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListUsersResponse represents a paginated list of users
type ListUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// UserResponse represents a standard API response
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(message string, data any) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(message string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
	}
}

// NewUserResponse creates a user response from repository user
func NewUserResponse(id, email, fullName, role, createdAt, updatedAt string) UserResponse {
	return UserResponse{
		ID:        id,
		Email:     email,
		FullName:  fullName,
		Role:      role,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}