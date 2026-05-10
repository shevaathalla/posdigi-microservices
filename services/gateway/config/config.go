package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	AuthServiceURL         string
	UserServiceURL         string
	AttendanceServiceURL   string
	JWTSecret              string
	RateLimitRequests      int
	RateLimitWindow        time.Duration
	InternalServiceKey     string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		if godotenv.Load("../../.env") != nil {
			log.Println("No .env file found, using default values")
		}
	}

	rateLimitWindow := getEnvDuration("RATE_LIMIT_WINDOW", 1*time.Minute)
	if rateLimitWindow == 0 {
		rateLimitWindow = 1 * time.Minute
	}

	return &Config{
		Port:                 getEnv("GATEWAY_PORT", "8000"),
		AuthServiceURL:       getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
		UserServiceURL:       getEnv("USER_SERVICE_URL", "http://localhost:8002"),
		AttendanceServiceURL: getEnv("ATTENDANCE_SERVICE_URL", "http://localhost:8003"),
		JWTSecret:            getEnv("JWT_SECRET", "crabbypatty"),
		RateLimitRequests:    getEnvInt("RATE_LIMIT_REQUESTS", 100),
		RateLimitWindow:      rateLimitWindow,
		InternalServiceKey:   getEnv("INTERNAL_SERVICE_KEY", "internal-service-key-change-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

// GetEnv retrieves environment variable with fallback (public utility)
func GetEnv(key, fallback string) string {
	return getEnv(key, fallback)
}
