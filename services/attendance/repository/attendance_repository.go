package repository

import (
	"errors"
	"posdigi-attendance/model"
	"time"

	"gorm.io/gorm"
)

// AttendanceRepository interface defines attendance database operations
type AttendanceRepository interface {
	Create(attendance *model.Attendance) error
	FindByID(id string) (*model.Attendance, error)
	FindActiveByUserID(userID string) (*model.Attendance, error)
	FindByUserID(userID string, limit, offset int) ([]*model.Attendance, error)
	CountByUserID(userID string) (int64, error)
	Update(attendance *model.Attendance) error
	GetTotalHours(userID string, startDate, endDate time.Time) (float64, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

// NewAttendanceRepository creates a new attendance repository instance
func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{db: db}
}

// Create creates a new attendance record
func (r *attendanceRepository) Create(attendance *model.Attendance) error {
	return r.db.Create(attendance).Error
}

// FindByID finds an attendance record by ID
func (r *attendanceRepository) FindByID(id string) (*model.Attendance, error) {
	var attendance model.Attendance
	err := r.db.Where("id = ?", id).First(&attendance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attendance, nil
}

// FindActiveByUserID finds an active attendance record (no clock out) for a user
func (r *attendanceRepository) FindActiveByUserID(userID string) (*model.Attendance, error) {
	var attendance model.Attendance
	err := r.db.Where("user_id = ? AND clock_out IS NULL", userID).
		Order("clock_in DESC").
		First(&attendance).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &attendance, nil
}

// FindByUserID finds attendance records for a user with pagination
func (r *attendanceRepository) FindByUserID(userID string, limit, offset int) ([]*model.Attendance, error) {
	var attendances []*model.Attendance
	err := r.db.Where("user_id = ?", userID).
		Order("clock_in DESC").
		Limit(limit).
		Offset(offset).
		Find(&attendances).Error
	if err != nil {
		return nil, err
	}
	return attendances, nil
}

// CountByUserID counts total attendance records for a user
func (r *attendanceRepository) CountByUserID(userID string) (int64, error) {
	var count int64
	err := r.db.Model(&model.Attendance{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// Update updates an attendance record
func (r *attendanceRepository) Update(attendance *model.Attendance) error {
	return r.db.Save(attendance).Error
}

// GetTotalHours calculates total work hours for a user within a date range
func (r *attendanceRepository) GetTotalHours(userID string, startDate, endDate time.Time) (float64, error) {
	var attendances []model.Attendance
	err := r.db.Where("user_id = ? AND clock_in >= ? AND clock_in <= ? AND clock_out IS NOT NULL",
		userID, startDate, endDate).
		Find(&attendances).Error
	if err != nil {
		return 0, err
	}

	var totalHours float64
	for _, attendance := range attendances {
		if attendance.ClockOut != nil {
			duration := attendance.ClockOut.Sub(attendance.ClockIn)
			totalHours += duration.Hours()
		}
	}

	return totalHours, nil
}
