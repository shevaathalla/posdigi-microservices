package dto

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email        string               `json:"email" validate:"required,email"`
	Password     string               `json:"password" validate:"required,min=6"`
	EmployeeData *EmployeeDataRequest `json:"employee_data,omitempty"`
}

// EmployeeDataRequest represents employee profile data during registration
type EmployeeDataRequest struct {
	FullName         string  `json:"full_name,omitempty" validate:"required,min=2,max=100"`
	Phone            string  `json:"phone,omitempty" validate:"omitempty,max=20"`
	Department       string  `json:"department,omitempty" validate:"omitempty,max=50"`
	Position         string  `json:"position,omitempty" validate:"omitempty,max=50"`
	Salary           float64 `json:"salary,omitempty" validate:"omitempty,min=0"`
	HireDate         string  `json:"hire_date,omitempty" validate:"omitempty"`
	EmploymentStatus string  `json:"employment_status,omitempty" validate:"omitempty,oneof=active terminated on_leave suspended"`
	ManagerID        *string `json:"manager_id,omitempty"`
	EmergencyContact string  `json:"emergency_contact,omitempty" validate:"omitempty,max=100"`
	EmergencyPhone   string  `json:"emergency_phone,omitempty" validate:"omitempty,max=20"`
	Address          string  `json:"address,omitempty" validate:"omitempty,max=500"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TokenRequest represents a token validation request
type TokenRequest struct {
	Token string `json:"token" validate:"required"`
}