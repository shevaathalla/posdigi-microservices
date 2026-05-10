package dto

// UserProfileResponse represents a user profile from User Service
type UserProfileResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// CreateUserResponse represents response from User Service when creating a user
type CreateUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}