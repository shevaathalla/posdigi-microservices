package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	InternalServiceKey string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		if godotenv.Load("../../.env") != nil {
			log.Println("No .env file found, using default values")
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
		Port:               getEnv("ATTENDANCE_PORT", "8003"),
		DatabaseURL:        getEnv("DATABASE_URL", databaseUrl),
		InternalServiceKey: getEnv("INTERNAL_SERVICE_KEY", "superpowers"),
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

// GetEnv retrieves environment variable with fallback (public utility)
func GetEnv(key, fallback string) string {
	return getEnv(key, fallback)
}
