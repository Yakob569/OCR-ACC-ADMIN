package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	DBUser        string
	DBPass        string
	DBHost        string
	DBPort        string
	DBName        string
	Port          string
	JWTSecret     string
	AdminUsername string
	AdminPassword string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	cfg := &Config{
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPass:        getEnv("DB_PASS", "postgres"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBName:        getEnv("DB_NAME", "postgres"),
		Port:          getEnv("PORT", "8082"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me-at-all-costs"),
		AdminUsername: getEnv("ADMIN_USERNAME", "admin"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "admin123"),
	}

	if cfg.DatabaseURL == "" && cfg.DBPass == "" {
		log.Println("Warning: Neither DATABASE_URL nor DB_PASS is set, using default values")
	}

	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		return nil, errors.New("ADMIN_USERNAME and ADMIN_PASSWORD must be configured")
	}

	if cfg.JWTSecret == "" || cfg.JWTSecret == "change-me-at-all-costs" {
		return nil, errors.New("JWT_SECRET must be configured securely")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	raw, ok := os.LookupEnv(key)
	if !ok || raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", key, err)
	}

	return value, nil
}
