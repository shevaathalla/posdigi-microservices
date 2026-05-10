package dto

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=user admin"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	FullName string `json:"full_name,omitempty" validate:"omitempty,min=2,max=100"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password,omitempty" validate:"omitempty,min=6"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=user admin"`
}

// ListUsersRequest represents a request to list users
type ListUsersRequest struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Search string `json:"search,omitempty"`
}

// Validate normalizes pagination defaults for ListUsersRequest.
// Field validation is handled by c.Validate() via CustomValidator.
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

// AuthenticateUserRequest represents an authentication request from the auth service
type AuthenticateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}