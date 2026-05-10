package dto

// ClockInRequest represents a request to clock in
type ClockInRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Notes  string `json:"notes,omitempty"`
}

// ClockOutRequest represents a request to clock out
type ClockOutRequest struct {
	AttendanceID string `json:"attendance_id" validate:"required"`
	Notes        string `json:"notes,omitempty"`
}

// GetAttendanceHistoryRequest represents a request to get attendance history
type GetAttendanceHistoryRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
}

// GetAttendanceSummaryRequest represents a request to get attendance summary
type GetAttendanceSummaryRequest struct {
	UserID    string `json:"user_id" validate:"required"`
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
}
