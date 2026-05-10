package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTExpiry   int
	MongoDB     *MongoDBConfig
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	Host           string
	Port           string
	DatabaseName   string
	Username       string
	Password       string
	AuthDB         string
	ConnectTimeout int
	PoolLimit      uint64
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		if godotenv.Load("../../.env") != nil {
			log.Fatalf("No .env file found: %v", err)
		}
	}

	databaseUrl := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	return &Config{
		Port:        GetEnv("AUTH_PORT", "8001"),
		DatabaseURL: GetEnv("DATABASE_URL", databaseUrl),
		JWTSecret:   GetEnv("JWT_SECRET", "crabbypatty"),
		JWTExpiry:   GetEnvInt("JWT_EXPIRY", 24),
		MongoDB: &MongoDBConfig{
			Host:           GetEnv("MONGODB_HOST", "localhost"),
			Port:           GetEnv("MONGODB_PORT", "27017"),
			DatabaseName:   GetEnv("MONGODB_DATABASE", "posdigi_activity_logs"),
			Username:       GetEnv("MONGODB_USERNAME", ""),
			Password:       GetEnv("MONGODB_PASSWORD", ""),
			AuthDB:         GetEnv("MONGODB_AUTH_DB", "admin"),
			ConnectTimeout: GetEnvInt("MONGODB_CONNECTION_TIMEOUT", 10),
			PoolLimit:      uint64(GetEnvInt("MONGODB_POOL_LIMIT", 100)),
		},
	}
}

func GetEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetEnv retrieves environment variable with fallback (public utility)
func GetEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
