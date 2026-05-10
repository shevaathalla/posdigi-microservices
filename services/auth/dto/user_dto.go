package dto

// CreateUserRequest represents a request to create a user profile
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

// CreateUserResponse represents the response from creating a user
type CreateUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// UserProfileResponse represents a user profile from User Service
type UserProfileResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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

// ListUsersResponse represents a paginated list of users
type ListUsersResponse struct {
	Users []UserProfileResponse `json:"users"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}