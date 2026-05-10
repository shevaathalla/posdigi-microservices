package dto

import (
	"errors"
	"strings"
	"time"
)

// ClockInRequest represents a request to clock in
type ClockInRequest struct {
	UserID string `json:"user_id"`
	Notes  string `json:"notes,omitempty"`
}

// ClockOutRequest represents a request to clock out
type ClockOutRequest struct {
	AttendanceID string `json:"attendance_id"`
	Notes        string `json:"notes,omitempty"`
}

// GetAttendanceHistoryRequest represents a request to get attendance history
type GetAttendanceHistoryRequest struct {
	UserID string `json:"user_id"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

// GetAttendanceSummaryRequest represents a request to get attendance summary
type GetAttendanceSummaryRequest struct {
	UserID    string `json:"user_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// Validate validates the clock in request
func (r *ClockInRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("User ID is required")
	}
	return nil
}

// Validate validates the clock out request
func (r *ClockOutRequest) Validate() error {
	if strings.TrimSpace(r.AttendanceID) == "" {
		return errors.New("Attendance ID is required")
	}
	return nil
}

// Validate validates the get attendance history request
func (r *GetAttendanceHistoryRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("User ID is required")
	}
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

// Validate validates the get attendance summary request
func (r *GetAttendanceSummaryRequest) Validate() error {
	if strings.TrimSpace(r.UserID) == "" {
		return errors.New("User ID is required")
	}
	if strings.TrimSpace(r.StartDate) == "" {
		return errors.New("Start date is required")
	}
	if strings.TrimSpace(r.EndDate) == "" {
		return errors.New("End date is required")
	}

	// Validate date format
	_, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return errors.New("Invalid start date format. Use YYYY-MM-DD")
	}
	_, err = time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return errors.New("Invalid end date format. Use YYYY-MM-DD")
	}

	return nil
}
