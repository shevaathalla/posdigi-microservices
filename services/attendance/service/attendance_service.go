package service

import (
	"errors"
	"fmt"
	"time"

	"posdigi-attendance/config"
	"posdigi-attendance/dto"
	"posdigi-attendance/model"
	"posdigi-attendance/repository"
)

type AttendanceService interface {
	ClockIn(req *dto.ClockInRequest) (*model.Attendance, error)
	ClockOut(req *dto.ClockOutRequest) (*model.Attendance, error)
	GetAttendanceHistory(req *dto.GetAttendanceHistoryRequest) (*dto.AttendanceHistoryResponse, error)
	GetAttendanceSummary(req *dto.GetAttendanceSummaryRequest) (*dto.AttendanceSummaryResponse, error)
}

type attendanceService struct {
	attendanceRepo repository.AttendanceRepository
	config         *config.Config
}

// NewAttendanceService creates a new attendance service instance
func NewAttendanceService(attendanceRepo repository.AttendanceRepository, cfg *config.Config) AttendanceService {
	return &attendanceService{
		attendanceRepo: attendanceRepo,
		config:         cfg,
	}
}

// ClockIn records a clock-in for a user
func (s *attendanceService) ClockIn(req *dto.ClockInRequest) (*model.Attendance, error) {
	config.Debug("Clocking in user: " + req.UserID)

	// Check if user has an active clock-in
	activeAttendance, err := s.attendanceRepo.FindActiveByUserID(req.UserID)
	if err != nil {
		config.Errorf("Database error checking active attendance: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if activeAttendance != nil {
		config.Warn("User already has an active clock-in: " + req.UserID)
		return nil, errors.New("user already has an active clock-in")
	}

	// Create new attendance record
	attendance := &model.Attendance{
		UserID:    req.UserID,
		ClockIn:   time.Now(),
		Notes:     req.Notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.attendanceRepo.Create(attendance); err != nil {
		config.Errorf("Error creating attendance: %v", err)
		return nil, fmt.Errorf("error creating attendance: %w", err)
	}

	config.Info("User clocked in successfully: " + req.UserID)
	return attendance, nil
}

// ClockOut records a clock-out for a user
func (s *attendanceService) ClockOut(req *dto.ClockOutRequest) (*model.Attendance, error) {
	config.Debug("Clocking out attendance: " + req.AttendanceID)

	// Find attendance record
	attendance, err := s.attendanceRepo.FindByID(req.AttendanceID)
	if err != nil {
		config.Errorf("Database error finding attendance: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	if attendance == nil {
		config.Warn("Attendance not found: " + req.AttendanceID)
		return nil, errors.New("attendance not found")
	}

	if attendance.ClockOut != nil {
		config.Warn("Attendance already clocked out: " + req.AttendanceID)
		return nil, errors.New("attendance already clocked out")
	}

	// Update clock-out time
	now := time.Now()
	attendance.ClockOut = &now
	attendance.UpdatedAt = time.Now()
	if req.Notes != "" {
		attendance.Notes = req.Notes
	}

	if err := s.attendanceRepo.Update(attendance); err != nil {
		config.Errorf("Error updating attendance: %v", err)
		return nil, fmt.Errorf("error updating attendance: %w", err)
	}

	config.Info("Attendance clocked out successfully: " + req.AttendanceID)
	return attendance, nil
}

// GetAttendanceHistory retrieves attendance history for a user
func (s *attendanceService) GetAttendanceHistory(req *dto.GetAttendanceHistoryRequest) (*dto.AttendanceHistoryResponse, error) {
	config.Debug("Getting attendance history for user: " + req.UserID)

	// Get total count
	total, err := s.attendanceRepo.CountByUserID(req.UserID)
	if err != nil {
		config.Errorf("Database error counting attendance: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Limit

	// Get attendance records
	attendances, err := s.attendanceRepo.FindByUserID(req.UserID, req.Limit, offset)
	if err != nil {
		config.Errorf("Database error finding attendance: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Convert to response format
	attendanceResponses := make([]dto.AttendanceResponse, 0, len(attendances))
	for _, attendance := range attendances {
		attendanceResponses = append(attendanceResponses, dto.NewAttendanceResponse(
			attendance.ID,
			attendance.UserID,
			attendance.ClockIn,
			attendance.ClockOut,
			attendance.Notes,
			attendance.CreatedAt.Format(time.RFC3339),
			attendance.UpdatedAt.Format(time.RFC3339),
		))
	}

	return &dto.AttendanceHistoryResponse{
		Attendances: attendanceResponses,
		Total:       int(total),
		Page:        req.Page,
		Limit:       req.Limit,
	}, nil
}

// GetAttendanceSummary retrieves attendance summary for a user within a date range
func (s *attendanceService) GetAttendanceSummary(req *dto.GetAttendanceSummaryRequest) (*dto.AttendanceSummaryResponse, error) {
	config.Debug("Getting attendance summary for user: " + req.UserID)

	// Parse dates
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, errors.New("invalid start date format")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, errors.New("invalid end date format")
	}

	// Set end date to end of day
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// Get total hours
	totalHours, err := s.attendanceRepo.GetTotalHours(req.UserID, startDate, endDate)
	if err != nil {
		config.Errorf("Database error calculating total hours: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Get count of attendance records for the period
	attendances, err := s.attendanceRepo.FindByUserID(req.UserID, 1000, 0)
	if err != nil {
		config.Errorf("Database error finding attendance: %v", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Filter attendance records within date range
	filteredAttendances := 0
	for _, attendance := range attendances {
		if (attendance.ClockIn.Equal(startDate) || attendance.ClockIn.After(startDate)) &&
			(attendance.ClockIn.Equal(endDate) || attendance.ClockIn.Before(endDate)) {
			filteredAttendances++
		}
	}

	// Calculate average hours
	averageHours := 0.0
	if filteredAttendances > 0 {
		averageHours = totalHours / float64(filteredAttendances)
	}

	return &dto.AttendanceSummaryResponse{
		UserID:       req.UserID,
		TotalHours:   totalHours,
		TotalDays:    filteredAttendances,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		AverageHours: averageHours,
	}, nil
}
