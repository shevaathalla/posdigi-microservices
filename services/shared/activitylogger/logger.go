package activitylogger

import (
	"context"
	"fmt"
	"time"
)

// Logger provides the main interface for activity logging
type Logger struct {
	repository *Repository
	service    string
}

// NewLogger creates a new activity logger for a specific service
func NewLogger(repository *Repository, serviceName string) *Logger {
	return &Logger{
		repository: repository,
		service:    serviceName,
	}
}

// LogActivity logs a single activity
func (l *Logger) LogActivity(ctx context.Context, log *ActivityLog) error {
	// Ensure service name is set
	if log.Service == "" {
		log.Service = l.service
	}

	// Set timestamps if not already set
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	return l.repository.Save(ctx, log)
}

// LogBatch logs multiple activities at once
func (l *Logger) LogBatch(ctx context.Context, logs []ActivityLog) error {
	// Ensure service names and timestamps are set
	now := time.Now()
	for i := range logs {
		if logs[i].Service == "" {
			logs[i].Service = l.service
		}
		if logs[i].Timestamp.IsZero() {
			logs[i].Timestamp = now
		}
		if logs[i].CreatedAt.IsZero() {
			logs[i].CreatedAt = now
		}
	}

	return l.repository.SaveBatch(ctx, logs)
}

// LogLogin logs user login attempts
func (l *Logger) LogLogin(ctx context.Context, requestID, userID, email, ip, userAgent string, success bool, errMsg string) error {
	action := ActionLoginSuccess
	if !success {
		action = ActionLoginFailed
	}

	log := NewActivityLog(l.service, action, requestID).
		WithUser(userID).
		WithRequest("POST", "/auth/login", ip, userAgent).
		WithResponse(200, success)

	if !success && errMsg != "" {
		log.ErrorMessage = errMsg
	}

	// Add metadata
	if log.Metadata == nil {
		log.Metadata = &ActivityMetadata{}
	}
	log.Metadata.Extra = map[string]interface{}{
		"email": email,
	}

	return l.LogActivity(ctx, log)
}

// LogLogout logs user logout events
func (l *Logger) LogLogout(ctx context.Context, requestID, userID, ip, userAgent string) error {
	log := NewActivityLog(l.service, ActionLogout, requestID).
		WithUser(userID).
		WithRequest("POST", "/auth/logout", ip, userAgent).
		WithResponse(200, true)

	return l.LogActivity(ctx, log)
}

// LogUserAction logs generic user CRUD operations
func (l *Logger) LogUserAction(ctx context.Context, requestID, userID, action, method, endpoint, ip, userAgent string, success bool, metadata *ActivityMetadata) error {
	log := NewActivityLog(l.service, action, requestID).
		WithUser(userID).
		WithRequest(method, endpoint, ip, userAgent).
		WithResponse(200, success).
		WithMetadata(metadata)

	if !success {
		log.ErrorMessage = fmt.Sprintf("Failed to %s", action)
	}

	return l.LogActivity(ctx, log)
}

// LogEmployeeAction logs employee-related operations
func (l *Logger) LogEmployeeAction(ctx context.Context, requestID, userID, employeeID, action, method, endpoint, ip, userAgent string, success bool, metadata *ActivityMetadata) error {
	log := NewActivityLog(l.service, action, requestID).
		WithUser(userID).
		WithEmployee(employeeID).
		WithRequest(method, endpoint, ip, userAgent).
		WithResponse(200, success).
		WithMetadata(metadata)

	if !success {
		log.ErrorMessage = fmt.Sprintf("Failed to %s", action)
	}

	return l.LogActivity(ctx, log)
}

// LogAttendanceAction logs attendance-related operations
func (l *Logger) LogAttendanceAction(ctx context.Context, requestID, userID, action, method, endpoint, ip, userAgent string, success bool, metadata *ActivityMetadata) error {
	log := NewActivityLog(l.service, action, requestID).
		WithUser(userID).
		WithRequest(method, endpoint, ip, userAgent).
		WithResponse(200, success).
		WithMetadata(metadata)

	if !success {
		log.ErrorMessage = fmt.Sprintf("Failed to %s", action)
	}

	return l.LogActivity(ctx, log)
}

// GetActivityHistory retrieves activity history for a user
func (l *Logger) GetActivityHistory(ctx context.Context, userID string, limit int64) ([]ActivityLog, error) {
	return l.repository.FindByUserID(ctx, userID, limit)
}

// GetEmployeeActivity retrieves activity history for an employee
func (l *Logger) GetEmployeeActivity(ctx context.Context, employeeID string, limit int64) ([]ActivityLog, error) {
	return l.repository.FindByEmployeeID(ctx, employeeID, limit)
}

// GetServiceActivity retrieves activity logs for this service
func (l *Logger) GetServiceActivity(ctx context.Context, limit int64) ([]ActivityLog, error) {
	return l.repository.FindByService(ctx, l.service, limit)
}

// GetStatistics retrieves activity statistics
func (l *Logger) GetStatistics(ctx context.Context) (*ActivityStats, error) {
	return l.repository.GetStatistics(ctx)
}

// QueryActivityLogs performs complex queries on activity logs
func (l *Logger) QueryActivityLogs(ctx context.Context, query ActivityLogQuery) ([]ActivityLog, error) {
	if query.Service == "" {
		query.Service = l.service
	}
	return l.repository.Find(ctx, query)
}

// GetByRequestID retrieves activity log by request ID
func (l *Logger) GetByRequestID(ctx context.Context, requestID string) (*ActivityLog, error) {
	return l.repository.FindByRequestID(ctx, requestID)
}

// CleanupOldLogs removes logs older than specified duration
func (l *Logger) CleanupOldLogs(ctx context.Context, duration time.Duration) (int64, error) {
	cutoffDate := time.Now().Add(-duration)
	return l.repository.DeleteOldLogs(ctx, cutoffDate)
}