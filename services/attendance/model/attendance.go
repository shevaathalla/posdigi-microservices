package model

import (
	"time"

	"gorm.io/gorm"
)

// Attendance represents the attendance model in the database
type Attendance struct {
	ID        string         `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID    string         `gorm:"type:uuid;not null;index" json:"user_id"`
	ClockIn   time.Time      `gorm:"not null" json:"clock_in"`
	ClockOut  *time.Time     `json:"clock_out,omitempty"`
	Notes     string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
