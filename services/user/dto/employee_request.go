package dto

// CreateEmployeeRequest represents the request to create a new employee
type CreateEmployeeRequest struct {
	UserID           string  `json:"user_id" binding:"required"`
	FullName         string  `json:"full_name" binding:"required,min=2,max=100"`
	Phone            string  `json:"phone" binding:"omitempty,max=20"`
	Department       string  `json:"department" binding:"omitempty,max=50"`
	Position         string  `json:"position" binding:"omitempty,max=50"`
	Salary           float64 `json:"salary" binding:"omitempty,min=0"`
	HireDate         string  `json:"hire_date" binding:"required"`
	EmploymentStatus string  `json:"employment_status" binding:"omitempty,oneof=active terminated on_leave suspended"`
	ManagerID        *string `json:"manager_id,omitempty"`
	EmergencyContact string  `json:"emergency_contact" binding:"omitempty,max=100"`
	EmergencyPhone   string  `json:"emergency_phone" binding:"omitempty,max=20"`
	Address          string  `json:"address" binding:"omitempty,max=500"`
}

// UpdateEmployeeRequest represents the request to update an employee
type UpdateEmployeeRequest struct {
	FullName         *string  `json:"full_name,omitempty"`
	Phone            *string  `json:"phone,omitempty"`
	Department       *string  `json:"department,omitempty"`
	Position         *string  `json:"position,omitempty"`
	Salary           *float64 `json:"salary,omitempty"`
	EmploymentStatus *string  `json:"employment_status,omitempty"`
	ManagerID        *string  `json:"manager_id,omitempty"`
	EmergencyContact *string  `json:"emergency_contact,omitempty"`
	EmergencyPhone   *string  `json:"emergency_phone,omitempty"`
	Address          *string  `json:"address,omitempty"`
}
