package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"posdigi-user/model"
)

// EmployeeRepository handles employee data operations
type EmployeeRepository struct {
	db *gorm.DB
}

// NewEmployeeRepository creates a new employee repository
func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

// Create creates a new employee record
func (r *EmployeeRepository) Create(employee *model.Employee) error {
	return r.db.Create(employee).Error
}

// GetByID retrieves an employee by ID
func (r *EmployeeRepository) GetByID(id string) (*model.Employee, error) {
	var employee model.Employee
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&employee).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetByUserID retrieves an employee by user ID
func (r *EmployeeRepository) GetByUserID(userID string) (*model.Employee, error) {
	var employee model.Employee
	err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).First(&employee).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetByEmployeeCode retrieves an employee by employee code
func (r *EmployeeRepository) GetByEmployeeCode(code string) (*model.Employee, error) {
	var employee model.Employee
	err := r.db.Where("employee_code = ? AND deleted_at IS NULL", code).First(&employee).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

// List retrieves all employees with pagination and filtering
func (r *EmployeeRepository) List(page, pageSize int, search string) ([]model.Employee, int64, error) {
	var employees []model.Employee
	var total int64

	query := r.db.Model(&model.Employee{}).Where("deleted_at IS NULL")

	// Search functionality
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("full_name ILIKE ? OR employee_code ILIKE ? OR department ILIKE ? OR position ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&employees).Error
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// Update updates an employee record
func (r *EmployeeRepository) Update(employee *model.Employee) error {
	return r.db.Save(employee).Error
}

// Delete performs a soft delete on an employee record
func (r *EmployeeRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&model.Employee{}).Error
}

// GetByDepartment retrieves employees by department
func (r *EmployeeRepository) GetByDepartment(department string, page, pageSize int) ([]model.Employee, int64, error) {
	var employees []model.Employee
	var total int64

	query := r.db.Model(&model.Employee{}).Where("department = ? AND deleted_at IS NULL", department)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("full_name ASC").Find(&employees).Error
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// GetByManager retrieves employees who report to a specific manager
func (r *EmployeeRepository) GetByManager(managerID string, page, pageSize int) ([]model.Employee, int64, error) {
	var employees []model.Employee
	var total int64

	query := r.db.Model(&model.Employee{}).Where("manager_id = ? AND deleted_at IS NULL", managerID)

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("full_name ASC").Find(&employees).Error
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// GetActiveEmployees retrieves all active employees
func (r *EmployeeRepository) GetActiveEmployees(page, pageSize int) ([]model.Employee, int64, error) {
	var employees []model.Employee
	var total int64

	query := r.db.Model(&model.Employee{}).Where("employment_status = ? AND deleted_at IS NULL", "active")

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("full_name ASC").Find(&employees).Error
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// GetEmployeeWithUser retrieves employee with associated user information
func (r *EmployeeRepository) GetEmployeeWithUser(id string) (*model.Employee, error) {
	var employee model.Employee
	err := r.db.Preload("User").Where("id = ? AND deleted_at IS NULL", id).First(&employee).Error
	if err != nil {
		return nil, err
	}
	return &employee, nil
}

// GetSubordinates retrieves all employees who report to a manager (recursive)
func (r *EmployeeRepository) GetSubordinates(managerID string) ([]model.Employee, error) {
	var employees []model.Employee
	err := r.db.Where("manager_id = ? AND deleted_at IS NULL", managerID).
		Order("full_name ASC").
		Find(&employees).Error
	if err != nil {
		return nil, err
	}
	return employees, nil
}

// UpdateEmploymentStatus updates the employment status of an employee
func (r *EmployeeRepository) UpdateEmploymentStatus(id, status string) error {
	return r.db.Model(&model.Employee{}).
		Where("id = ?", id).
		Update("employment_status", status).Error
}

// ExistsByEmployeeCode checks if an employee code already exists
func (r *EmployeeRepository) ExistsByEmployeeCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Employee{}).
		Where("employee_code = ? AND deleted_at IS NULL", code).
		Count(&count).Error
	return count > 0, err
}

// ExistsByUserID checks if an employee profile exists for a user
func (r *EmployeeRepository) ExistsByUserID(userID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Employee{}).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Count(&count).Error
	return count > 0, err
}

// GenerateNextEmployeeCode generates the next available employee code
func (r *EmployeeRepository) GenerateNextEmployeeCode() (string, error) {
	var lastEmployee model.Employee
	err := r.db.Order("employee_code DESC").First(&lastEmployee).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "EMP001", nil
		}
		return "", err
	}

	// Extract numeric part and increment
	lastCode := lastEmployee.EmployeeCode
	var lastNum int
	_, err = fmt.Sscanf(lastCode, "EMP%d", &lastNum)
	if err != nil {
		return "EMP001", nil
	}

	nextNum := lastNum + 1
	return fmt.Sprintf("EMP%03d", nextNum), nil
}
