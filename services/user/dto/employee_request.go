package dto

// CreateEmployeeRequest represents the request to create a new employee
type CreateEmployeeRequest struct {
	UserID           string  `json:"user_id" validate:"required"`
	FullName         string  `json:"full_name" validate:"required,min=2,max=100"`
	Phone            string  `json:"phone" validate:"omitempty,max=20"`
	Department       string  `json:"department" validate:"omitempty,max=50"`
	Position         string  `json:"position" validate:"omitempty,max=50"`
	Salary           float64 `json:"salary" validate:"omitempty,min=0"`
	HireDate         string  `json:"hire_date" validate:"required"`
	EmploymentStatus string  `json:"employment_status" validate:"omitempty,oneof=active terminated on_leave suspended"`
	ManagerID        *string `json:"manager_id,omitempty"`
	EmergencyContact string  `json:"emergency_contact" validate:"omitempty,max=100"`
	EmergencyPhone   string  `json:"emergency_phone" validate:"omitempty,max=20"`
	Address          string  `json:"address" validate:"omitempty,max=500"`
}

// UpdateEmployeeRequest represents the request to update an employee
type UpdateEmployeeRequest struct {
	FullName         *string  `json:"full_name,omitempty" validate:"omitempty,min=2,max=100"`
	Phone            *string  `json:"phone,omitempty" validate:"omitempty,max=20"`
	Department       *string  `json:"department,omitempty" validate:"omitempty,max=50"`
	Position         *string  `json:"position,omitempty" validate:"omitempty,max=50"`
	Salary           *float64 `json:"salary,omitempty" validate:"omitempty,min=0"`
	EmploymentStatus *string  `json:"employment_status,omitempty" validate:"omitempty,oneof=active terminated on_leave suspended"`
	ManagerID        *string  `json:"manager_id,omitempty"`
	EmergencyContact *string  `json:"emergency_contact,omitempty" validate:"omitempty,max=100"`
	EmergencyPhone   *string  `json:"emergency_phone,omitempty" validate:"omitempty,max=20"`
	Address          *string  `json:"address,omitempty" validate:"omitempty,max=500"`
}
