// MongoDB Initialization Script
// This script runs automatically when MongoDB container starts for the first time

// Get or create database
db = db.getSiblingDB('posdigi_activity_logs');

// Create activity_logs collection
db.createCollection('activity_logs');

// Create indexes for performance
db.activity_logs.createIndex({ "user_id": 1, "timestamp": -1 });
db.activity_logs.createIndex({ "service": 1, "timestamp": -1 });
db.activity_logs.createIndex({ "action": 1, "timestamp": -1 });
db.activity_logs.createIndex({ "request_id": 1 }, { unique: true });
db.activity_logs.createIndex({ "employee_id": 1, "timestamp": -1 });
db.activity_logs.createIndex({ "user_id": 1, "action": 1, "timestamp": -1 });

// Create text index for searching
db.activity_logs.createIndex({
  "action": "text",
  "endpoint": "text",
  "error_message": "text"
});

// Print success message
print('MongoDB initialization completed successfully!');
print('Created activity_logs collection with indexes');
print('Database: posdigi_activity_logs');