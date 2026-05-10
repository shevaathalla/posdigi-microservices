package activitylogger

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityLog represents a single activity log entry
type ActivityLog struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	UserID      string             `json:"user_id,omitempty" bson:"user_id,omitempty"`
	EmployeeID  string             `json:"employee_id,omitempty" bson:"employee_id,omitempty"`
	Service     string             `json:"service" bson:"service"`
	Action      string             `json:"action" bson:"action"`
	Endpoint    string             `json:"endpoint,omitempty" bson:"endpoint,omitempty"`
	Method      string             `json:"method,omitempty" bson:"method,omitempty"`
	IPAddress   string             `json:"ip_address,omitempty" bson:"ip_address,omitempty"`
	UserAgent   string             `json:"user_agent,omitempty" bson:"user_agent,omitempty"`
	RequestID   string             `json:"request_id" bson:"request_id"`
	StatusCode  int                `json:"status_code,omitempty" bson:"status_code,omitempty"`
	Success     bool               `json:"success" bson:"success"`
	ErrorMessage string            `json:"error_message,omitempty" bson:"error_message,omitempty"`
	Metadata    *ActivityMetadata  `json:"metadata,omitempty" bson:"metadata,omitempty"`
	Timestamp   time.Time          `json:"timestamp" bson:"timestamp"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}

// ActivityMetadata holds additional flexible data
type ActivityMetadata struct {
	Before  map[string]interface{} `json:"before,omitempty" bson:"before,omitempty"`
	After   map[string]interface{} `json:"after,omitempty" bson:"after,omitempty"`
	Changes []string               `json:"changes,omitempty" bson:"changes,omitempty"`
	Extra   map[string]interface{} `json:"extra,omitempty" bson:"extra,omitempty"`
}

// ActionType defines common activity actions
const (
	// Authentication actions
	ActionLoginSuccess    = "LOGIN_SUCCESS"
	ActionLoginFailed     = "LOGIN_FAILED"
	ActionLogout          = "LOGOUT"
	ActionRegister        = "REGISTER"
	ActionTokenGenerated  = "TOKEN_GENERATED"
	ActionTokenValidated  = "TOKEN_VALIDATED"
	ActionTokenRefreshed  = "TOKEN_REFRESHED"

	// User management actions
	ActionUserCreated     = "USER_CREATED"
	ActionUserUpdated     = "USER_UPDATED"
	ActionUserDeleted     = "USER_DELETED"
	ActionUserViewed      = "USER_VIEWED"

	// Employee actions
	ActionEmployeeCreated = "EMPLOYEE_CREATED"
	ActionEmployeeUpdated = "EMPLOYEE_UPDATED"
	ActionEmployeeDeleted = "EMPLOYEE_DELETED"
	ActionEmployeeViewed  = "EMPLOYEE_VIEWED"

	// Attendance actions
	ActionClockIn         = "CLOCK_IN"
	ActionClockOut        = "CLOCK_OUT"
	ActionAttendanceViewed = "ATTENDANCE_VIEWED"
	ActionAttendanceUpdated = "ATTENDANCE_UPDATED"

	// System actions
	ActionHealthCheck     = "HEALTH_CHECK"
	ActionMigrationRun    = "MIGRATION_RUN"
	ActionConfigChanged   = "CONFIG_CHANGED"
)

// ServiceName defines service identifiers
const (
	ServiceGateway     = "gateway"
	ServiceAuth        = "auth"
	ServiceUser        = "user"
	ServiceAttendance  = "attendance"
)

// ActivityLogQuery defines filters for querying activity logs
type ActivityLogQuery struct {
	UserID     string
	EmployeeID string
	Service    string
	Action     string
	StartDate  *time.Time
	EndDate    *time.Time
	Success    *bool
	Limit       int64
	Skip        int64
}

// ActivityStats represents activity statistics
type ActivityStats struct {
	TotalLogs      int64                    `json:"total_logs"`
	ByService      map[string]int64         `json:"by_service"`
	ByAction       map[string]int64         `json:"by_action"`
	ByUser         map[string]int64         `json:"by_user"`
	SuccessRate    float64                  `json:"success_rate"`
	TopUsers       []UserActivitySummary    `json:"top_users"`
	TimeDistribution map[string]int64       `json:"time_distribution"`
}

// UserActivitySummary shows activity summary for a user
type UserActivitySummary struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email,omitempty"`
	Count     int64  `json:"count"`
	LastLogin string `json:"last_login,omitempty"`
}

// NewActivityLog creates a new activity log entry with defaults
func NewActivityLog(service, action, requestID string) *ActivityLog {
	now := time.Now()
	return &ActivityLog{
		ID:        primitive.NewObjectID(),
		Service:   service,
		Action:    action,
		RequestID: requestID,
		Success:   true,
		Timestamp: now,
		CreatedAt: now,
	}
}

// WithUser sets user information for the activity log
func (al *ActivityLog) WithUser(userID string) *ActivityLog {
	al.UserID = userID
	return al
}

// WithEmployee sets employee information for the activity log
func (al *ActivityLog) WithEmployee(employeeID string) *ActivityLog {
	al.EmployeeID = employeeID
	return al
}

// WithRequest sets HTTP request information
func (al *ActivityLog) WithRequest(method, endpoint, ip, userAgent string) *ActivityLog {
	al.Method = method
	al.Endpoint = endpoint
	al.IPAddress = ip
	al.UserAgent = userAgent
	return al
}

// WithResponse sets response information
func (al *ActivityLog) WithResponse(statusCode int, success bool) *ActivityLog {
	al.StatusCode = statusCode
	al.Success = success
	return al
}

// WithError sets error information
func (al *ActivityLog) WithError(err error) *ActivityLog {
	al.Success = false
	if err != nil {
		al.ErrorMessage = err.Error()
	}
	return al
}

// WithMetadata sets metadata for the activity log
func (al *ActivityLog) WithMetadata(metadata *ActivityMetadata) *ActivityLog {
	al.Metadata = metadata
	return al
}