package model

import (
	"time"

	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Employee represents detailed employee information
type Employee struct {
	ID               string  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID           string  `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	EmployeeCode     string  `gorm:"type:varchar(20);uniqueIndex;not null" json:"employee_code"`
	FullName         string  `gorm:"type:varchar(100);not null" json:"full_name"`
	Phone            string  `gorm:"type:varchar(20)" json:"phone"`
	Department       string  `gorm:"type:varchar(50)" json:"department"`
	Position         string  `gorm:"type:varchar(50)" json:"position"`
	Salary           float64 `gorm:"type:decimal(10,2)" json:"salary"`
	HireDate         time.Time `gorm:"type:date;not null" json:"hire_date"`
	EmploymentStatus string  `gorm:"type:varchar(20);default:'active'" json:"employment_status"`
	ManagerID        *string `gorm:"type:uuid" json:"manager_id,omitempty"`
	EmergencyContact string  `gorm:"type:varchar(100)" json:"emergency_contact"`
	EmergencyPhone   string  `gorm:"type:varchar(20)" json:"emergency_phone"`
	Address          string  `gorm:"type:text" json:"address"`
	ProfileImage     string  `gorm:"type:varchar(255)" json:"profile_image,omitempty"`
	Metadata         JSON    `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	DeletedAt        *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	User    User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Manager *Employee `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`
}

// JSON custom type for storing JSON data in PostgreSQL
type JSON json.RawMessage

// Scan implements the sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSON value")
	}
	*j = JSON(bytes)
	return nil
}

// Value implements the driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// TableName specifies the table name for Employee model
func (Employee) TableName() string {
	return "employees"
}
