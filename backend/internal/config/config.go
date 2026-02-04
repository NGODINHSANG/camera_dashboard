package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabasePath string
	JWTSecret    string
	CORSOrigins  string
}

func Load() *Config {
	godotenv.Load()

	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./data/dashboard.db"),
		JWTSecret:    getEnv("JWT_SECRET", "default-secret-key"),
		CORSOrigins:  getEnv("CORS_ORIGINS", "http://localhost:5173"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
