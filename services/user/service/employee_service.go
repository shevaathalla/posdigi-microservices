package service

import (
	"errors"
	"time"

	"posdigi-user/model"
	"posdigi-user/repository"
)

// EmployeeService handles employee business logic
type EmployeeService struct {
	employeeRepo *repository.EmployeeRepository
	userRepo     repository.UserRepository
}

// NewEmployeeService creates a new employee service
func NewEmployeeService(employeeRepo *repository.EmployeeRepository, userRepo repository.UserRepository) *EmployeeService {
	return &EmployeeService{
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
	}
}

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

// CreateEmployee creates a new employee profile
func (s *EmployeeService) CreateEmployee(req *CreateEmployeeRequest) (*model.Employee, error) {
	// Validate user exists
	user, err := s.userRepo.FindByID(req.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Check if employee profile already exists for this user
	exists, err := s.employeeRepo.ExistsByUserID(req.UserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("employee profile already exists for this user")
	}

	// Validate manager exists if provided
	if req.ManagerID != nil && *req.ManagerID != "" {
		_, err := s.employeeRepo.GetByID(*req.ManagerID)
		if err != nil {
			return nil, errors.New("manager not found")
		}
	}

	// Parse hire date
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return nil, errors.New("invalid hire date format, use YYYY-MM-DD")
	}

	// Generate employee code
	employeeCode, err := s.employeeRepo.GenerateNextEmployeeCode()
	if err != nil {
		return nil, err
	}

	// Set default employment status if not provided
	employmentStatus := req.EmploymentStatus
	if employmentStatus == "" {
		employmentStatus = "active"
	}

	// Create employee
	employee := &model.Employee{
		UserID:           req.UserID,
		EmployeeCode:     employeeCode,
		FullName:         req.FullName,
		Phone:            req.Phone,
		Department:       req.Department,
		Position:         req.Position,
		Salary:           req.Salary,
		HireDate:         hireDate,
		EmploymentStatus: employmentStatus,
		ManagerID:        req.ManagerID,
		EmergencyContact: req.EmergencyContact,
		EmergencyPhone:   req.EmergencyPhone,
		Address:          req.Address,
	}

	err = s.employeeRepo.Create(employee)
	if err != nil {
		return nil, err
	}

	// Fetch the created employee with relations
	return s.employeeRepo.GetEmployeeWithUser(employee.ID)
}

// GetEmployee retrieves an employee by ID
func (s *EmployeeService) GetEmployee(id string) (*model.Employee, error) {
	employee, err := s.employeeRepo.GetEmployeeWithUser(id)
	if err != nil {
		return nil, errors.New("employee not found")
	}
	return employee, nil
}

// GetEmployeeByUserID retrieves an employee by user ID
func (s *EmployeeService) GetEmployeeByUserID(userID string) (*model.Employee, error) {
	employee, err := s.employeeRepo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("employee profile not found for this user")
	}
	return employee, nil
}

// GetEmployeeByCode retrieves an employee by employee code
func (s *EmployeeService) GetEmployeeByCode(code string) (*model.Employee, error) {
	employee, err := s.employeeRepo.GetByEmployeeCode(code)
	if err != nil {
		return nil, errors.New("employee not found")
	}
	return employee, nil
}

// ListEmployees retrieves all employees with pagination
func (s *EmployeeService) ListEmployees(page, pageSize int, search string) ([]model.Employee, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.employeeRepo.List(page, pageSize, search)
}

// UpdateEmployee updates an employee profile
func (s *EmployeeService) UpdateEmployee(id string, req *UpdateEmployeeRequest) (*model.Employee, error) {
	// Get existing employee
	employee, err := s.employeeRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("employee not found")
	}

	// Validate manager exists if provided
	if req.ManagerID != nil && *req.ManagerID != "" && *req.ManagerID != employee.ID {
		_, err := s.employeeRepo.GetByID(*req.ManagerID)
		if err != nil {
			return nil, errors.New("manager not found")
		}
	}

	// Update fields if provided
	if req.FullName != nil {
		employee.FullName = *req.FullName
	}
	if req.Phone != nil {
		employee.Phone = *req.Phone
	}
	if req.Department != nil {
		employee.Department = *req.Department
	}
	if req.Position != nil {
		employee.Position = *req.Position
	}
	if req.Salary != nil {
		employee.Salary = *req.Salary
	}
	if req.EmploymentStatus != nil {
		employee.EmploymentStatus = *req.EmploymentStatus
	}
	if req.ManagerID != nil {
		employee.ManagerID = req.ManagerID
	}
	if req.EmergencyContact != nil {
		employee.EmergencyContact = *req.EmergencyContact
	}
	if req.EmergencyPhone != nil {
		employee.EmergencyPhone = *req.EmergencyPhone
	}
	if req.Address != nil {
		employee.Address = *req.Address
	}

	err = s.employeeRepo.Update(employee)
	if err != nil {
		return nil, err
	}

	return s.employeeRepo.GetEmployeeWithUser(id)
}

// DeleteEmployee deletes an employee profile
func (s *EmployeeService) DeleteEmployee(id string) error {
	_, err := s.employeeRepo.GetByID(id)
	if err != nil {
		return errors.New("employee not found")
	}

	return s.employeeRepo.Delete(id)
}

// GetEmployeesByDepartment retrieves employees by department
func (s *EmployeeService) GetEmployeesByDepartment(department string, page, pageSize int) ([]model.Employee, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.employeeRepo.GetByDepartment(department, page, pageSize)
}

// GetSubordinates retrieves employees who report to a manager
func (s *EmployeeService) GetSubordinates(managerID string, page, pageSize int) ([]model.Employee, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Verify manager exists
	_, err := s.employeeRepo.GetByID(managerID)
	if err != nil {
		return nil, 0, errors.New("manager not found")
	}

	return s.employeeRepo.GetByManager(managerID, page, pageSize)
}

// GetActiveEmployees retrieves all active employees
func (s *EmployeeService) GetActiveEmployees(page, pageSize int) ([]model.Employee, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	return s.employeeRepo.GetActiveEmployees(page, pageSize)
}

// UpdateEmploymentStatus updates the employment status of an employee
func (s *EmployeeService) UpdateEmploymentStatus(id, status string) error {
	validStatuses := map[string]bool{
		"active":     true,
		"terminated": true,
		"on_leave":   true,
		"suspended":  true,
	}

	if !validStatuses[status] {
		return errors.New("invalid employment status")
	}

	_, err := s.employeeRepo.GetByID(id)
	if err != nil {
		return errors.New("employee not found")
	}

	return s.employeeRepo.UpdateEmploymentStatus(id, status)
}

// GetEmployeeOrganizationChart retrieves the organizational hierarchy
func (s *EmployeeService) GetEmployeeOrganizationChart(employeeID string) (*model.Employee, error) {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return nil, errors.New("employee not found")
	}

	// Load manager hierarchy
	if employee.ManagerID != nil {
		manager, err := s.getManagerHierarchy(*employee.ManagerID)
		if err == nil {
			employee.Manager = manager
		}
	}

	return employee, nil
}

// getManagerHierarchy recursively loads manager hierarchy
func (s *EmployeeService) getManagerHierarchy(managerID string) (*model.Employee, error) {
	manager, err := s.employeeRepo.GetByID(managerID)
	if err != nil {
		return nil, err
	}

	if manager.ManagerID != nil {
		manager.Manager, _ = s.getManagerHierarchy(*manager.ManagerID)
	}

	return manager, nil
}

// GetAllSubordinates retrieves all subordinates recursively
func (s *EmployeeService) GetAllSubordinates(managerID string) ([]model.Employee, error) {
	directReports, err := s.employeeRepo.GetSubordinates(managerID)
	if err != nil {
		return nil, err
	}

	// Recursively get subordinates of subordinates
	var allSubordinates []model.Employee
	for _, employee := range directReports {
		allSubordinates = append(allSubordinates, employee)

		if employee.ID != "" {
			subSubordinates, _ := s.GetAllSubordinates(employee.ID)
			allSubordinates = append(allSubordinates, subSubordinates...)
		}
	}

	return allSubordinates, nil
}

// GetEmployeeProfile retrieves complete employee profile with user info
func (s *EmployeeService) GetEmployeeProfile(id string) (map[string]interface{}, error) {
	employee, err := s.employeeRepo.GetEmployeeWithUser(id)
	if err != nil {
		return nil, errors.New("employee not found")
	}

	// Remove sensitive user data
	if employee.User.Password != "" {
		employee.User.Password = ""
	}

	// Build response
	response := map[string]interface{}{
		"employee": employee,
		"user": map[string]interface{}{
			"id":         employee.User.ID,
			"email":      employee.User.Email,
			"created_at": employee.User.CreatedAt,
		},
	}

	return response, nil
}
