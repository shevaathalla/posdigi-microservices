package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config holds MongoDB configuration
type Config struct {
	Host            string
	Port            string
	DatabaseName    string
	Username        string
	Password        string
	AuthDB          string
	ConnectTimeout  int
	PoolLimit       uint64
}

// DefaultConfig returns default MongoDB configuration
func DefaultConfig() Config {
	return Config{
		Host:            "localhost",
		Port:            "27017",
		DatabaseName:    "posdigi_activity_logs",
		AuthDB:          "admin",
		ConnectTimeout:  10,
		PoolLimit:       100,
	}
}

// Client wraps MongoDB client with additional functionality
type Client struct {
	*mongo.Client
	database *mongo.Database
	config   Config
}

// ConnectMongoDB establishes connection to MongoDB
func ConnectMongoDB(cfg Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ConnectTimeout)*time.Second)
	defer cancel()

	// Build connection URI
	uri := buildConnectionURI(cfg)

	// Configure client options
	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(cfg.PoolLimit)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	mongoClient := &Client{
		Client:   client,
		database: client.Database(cfg.DatabaseName),
		config:   cfg,
	}

	log.Printf("Successfully connected to MongoDB at %s", uri)

	return mongoClient, nil
}

// buildConnectionURI constructs MongoDB connection URI
func buildConnectionURI(cfg Config) string {
	if cfg.Username != "" && cfg.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%s/%s?authSource=%s",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DatabaseName, cfg.AuthDB)
	}
	return fmt.Sprintf("mongodb://%s:%s/%s", cfg.Host, cfg.Port, cfg.DatabaseName)
}

// Database returns the MongoDB database instance
func (c *Client) Database() *mongo.Database {
	return c.database
}

// Collection returns a specific collection from the database
func (c *Client) Collection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

// Close closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	return c.Client.Disconnect(ctx)
}

// Ping checks if MongoDB connection is alive
func (c *Client) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx, nil)
}