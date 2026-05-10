package activitylogger

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	activityLogsCollection = "activity_logs"
)

// Repository handles MongoDB operations for activity logs
type Repository struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewRepository creates a new activity log repository
func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		db: db,
	}
}

// Collection returns the activity logs collection
func (r *Repository) Collection() *mongo.Collection {
	return r.db.Collection(activityLogsCollection)
}

// Save logs a single activity entry
func (r *Repository) Save(ctx context.Context, log *ActivityLog) error {
	if log.ID.IsZero() {
		log.ID = mongoInsertID()
	}

	_, err := r.Collection().InsertOne(ctx, log)
	if err != nil {
		return err
	}

	return nil
}

// SaveBatch logs multiple activity entries in bulk
func (r *Repository) SaveBatch(ctx context.Context, logs []ActivityLog) error {
	if len(logs) == 0 {
		return nil
	}

	var documents []interface{}
	for i := range logs {
		if logs[i].ID.IsZero() {
			logs[i].ID = mongoInsertID()
		}
		documents = append(documents, logs[i])
	}

	_, err := r.Collection().InsertMany(ctx, documents)
	if err != nil {
		return err
	}

	return nil
}

// FindByID retrieves an activity log by ID
func (r *Repository) FindByID(ctx context.Context, id string) (*ActivityLog, error) {
	objectID, err := toObjectID(id)
	if err != nil {
		return nil, err
	}

	var log ActivityLog
	err = r.Collection().FindOne(ctx, bson.M{"_id": objectID}).Decode(&log)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("activity log not found")
		}
		return nil, err
	}

	return &log, nil
}

// Find retrieves activity logs based on query parameters
func (r *Repository) Find(ctx context.Context, query ActivityLogQuery) ([]ActivityLog, error) {
	mongoQuery := r.buildQuery(query)

	opts := options.Find()
	if query.Limit > 0 {
		opts.SetLimit(query.Limit)
	}
	if query.Skip > 0 {
		opts.SetSkip(query.Skip)
	}

	// Sort by timestamp descending (newest first)
	opts.SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := r.Collection().Find(ctx, mongoQuery, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []ActivityLog
	if err = cursor.All(ctx, &logs); err != nil {
		return nil, err
	}

	return logs, nil
}

// FindByUserID retrieves all activity logs for a specific user
func (r *Repository) FindByUserID(ctx context.Context, userID string, limit int64) ([]ActivityLog, error) {
	query := ActivityLogQuery{
		UserID: userID,
		Limit:  limit,
	}
	return r.Find(ctx, query)
}

// FindByEmployeeID retrieves all activity logs for a specific employee
func (r *Repository) FindByEmployeeID(ctx context.Context, employeeID string, limit int64) ([]ActivityLog, error) {
	query := ActivityLogQuery{
		EmployeeID: employeeID,
		Limit:      limit,
	}
	return r.Find(ctx, query)
}

// FindByService retrieves activity logs for a specific service
func (r *Repository) FindByService(ctx context.Context, service string, limit int64) ([]ActivityLog, error) {
	query := ActivityLogQuery{
		Service: service,
		Limit:   limit,
	}
	return r.Find(ctx, query)
}

// FindByAction retrieves activity logs for a specific action type
func (r *Repository) FindByAction(ctx context.Context, action string, limit int64) ([]ActivityLog, error) {
	query := ActivityLogQuery{
		Action: action,
		Limit:  limit,
	}
	return r.Find(ctx, query)
}

// FindByRequestID retrieves activity log by request ID
func (r *Repository) FindByRequestID(ctx context.Context, requestID string) (*ActivityLog, error) {
	var log ActivityLog
	err := r.Collection().FindOne(ctx, bson.M{"request_id": requestID}).Decode(&log)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("activity log not found")
		}
		return nil, err
	}

	return &log, nil
}

// Count returns the total count of activity logs matching the query
func (r *Repository) Count(ctx context.Context, query ActivityLogQuery) (int64, error) {
	mongoQuery := r.buildQuery(query)
	return r.Collection().CountDocuments(ctx, mongoQuery)
}

// GetStatistics returns activity statistics
func (r *Repository) GetStatistics(ctx context.Context) (*ActivityStats, error) {
	stats := &ActivityStats{
		ByService:       make(map[string]int64),
		ByAction:        make(map[string]int64),
		ByUser:          make(map[string]int64),
		TimeDistribution: make(map[string]int64),
	}

	// Total count
	total, err := r.Collection().CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	stats.TotalLogs = total

	// Group by service
	servicePipeline := []bson.M{
		{"$group": bson.M{"_id": "$service", "count": bson.M{"$sum": 1}}},
	}
	serviceCursor, err := r.Collection().Aggregate(ctx, servicePipeline)
	if err == nil {
		var results []bson.M
		serviceCursor.All(ctx, &results)
		for _, result := range results {
			service := result["_id"].(string)
			count := result["count"].(int32)
			stats.ByService[service] = int64(count)
		}
		serviceCursor.Close(ctx)
	}

	// Group by action
	actionPipeline := []bson.M{
		{"$group": bson.M{"_id": "$action", "count": bson.M{"$sum": 1}}},
	}
	actionCursor, err := r.Collection().Aggregate(ctx, actionPipeline)
	if err == nil {
		var results []bson.M
		actionCursor.All(ctx, &results)
		for _, result := range results {
			action := result["_id"].(string)
			count := result["count"].(int32)
			stats.ByAction[action] = int64(count)
		}
		actionCursor.Close(ctx)
	}

	// Calculate success rate
	successCount, err := r.Collection().CountDocuments(ctx, bson.M{"success": true})
	if err == nil && stats.TotalLogs > 0 {
		stats.SuccessRate = float64(successCount) / float64(stats.TotalLogs) * 100
	}

	return stats, nil
}

// DeleteOldLogs deletes activity logs older than specified duration
func (r *Repository) DeleteOldLogs(ctx context.Context, olderThanDate interface{}) (int64, error) {
	result, err := r.Collection().DeleteMany(ctx, bson.M{"timestamp": bson.M{"$lt": olderThanDate}})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// buildQuery constructs MongoDB query from ActivityLogQuery
func (r *Repository) buildQuery(query ActivityLogQuery) bson.M {
	mongoQuery := bson.M{}

	if query.UserID != "" {
		mongoQuery["user_id"] = query.UserID
	}

	if query.EmployeeID != "" {
		mongoQuery["employee_id"] = query.EmployeeID
	}

	if query.Service != "" {
		mongoQuery["service"] = query.Service
	}

	if query.Action != "" {
		mongoQuery["action"] = query.Action
	}

	if query.Success != nil {
		mongoQuery["success"] = *query.Success
	}

	// Date range filter
	if query.StartDate != nil || query.EndDate != nil {
		timeFilter := bson.M{}
		if query.StartDate != nil {
			timeFilter["$gte"] = query.StartDate
		}
		if query.EndDate != nil {
			timeFilter["$lte"] = query.EndDate
		}
		mongoQuery["timestamp"] = timeFilter
	}

	return mongoQuery
}

// Helper functions
func mongoInsertID() primitive.ObjectID {
	return primitive.NewObjectID()
}

func toObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}