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
		Port:        getEnv("AUTH_PORT", "8001"),
		DatabaseURL: getEnv("DATABASE_URL", databaseUrl),
		JWTSecret:   getEnv("JWT_SECRET", "crabbypatty"),
		JWTExpiry:   getEnvInt("JWT_EXPIRY", 24),
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
