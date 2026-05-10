package dto

import "time"

// AttendanceResponse represents an attendance record in responses
type AttendanceResponse struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	ClockIn   string     `json:"clock_in"`
	ClockOut  *string    `json:"clock_out,omitempty"`
	Notes     string     `json:"notes,omitempty"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
}

// AttendanceHistoryResponse represents a paginated list of attendance records
type AttendanceHistoryResponse struct {
	Attendances []AttendanceResponse `json:"attendances"`
	Total       int                  `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
}

// AttendanceSummaryResponse represents attendance summary for a user
type AttendanceSummaryResponse struct {
	UserID         string  `json:"user_id"`
	TotalHours     float64 `json:"total_hours"`
	TotalDays      int     `json:"total_days"`
	StartDate      string  `json:"start_date"`
	EndDate        string  `json:"end_date"`
	AverageHours   float64 `json:"average_hours"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// NewSuccessResponse creates a success response
func NewSuccessResponse(message string, data any) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse creates an error response
func NewErrorResponse(message string) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
	}
}

// NewAttendanceResponse creates an attendance response from repository attendance
func NewAttendanceResponse(id, userID string, clockIn time.Time, clockOut *time.Time, notes, createdAt, updatedAt string) AttendanceResponse {
	var clockOutStr *string
	if clockOut != nil {
		formatted := clockOut.Format(time.RFC3339)
		clockOutStr = &formatted
	}

	return AttendanceResponse{
		ID:        id,
		UserID:    userID,
		ClockIn:   clockIn.Format(time.RFC3339),
		ClockOut:  clockOutStr,
		Notes:     notes,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
