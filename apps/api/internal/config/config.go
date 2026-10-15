package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configurations for the SMARTREVERSAL API service.
type Config struct {
	Environment  string
	Port         int
	DatabaseURL  string
	DBHost       string
	DBPort       int
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	RedisHost    string
	RedisPort    int
	JWTSecret    string
	AIServiceURL string
}

// Load reads configuration from environment variables with safe development defaults.
func Load() *Config {
	port, _ := strconv.Atoi(getEnv("API_PORT", "8080"))
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))

	return &Config{
		Environment:  getEnv("ENVIRONMENT", "development"),
		Port:         port,
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       dbPort,
		DBUser:       getEnv("DB_USER", "smartreversal_app"),
		DBPassword:   getEnv("DB_PASSWORD", "development_secret_only"),
		DBName:       getEnv("DB_NAME", "smartreversal_db"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		RedisHost:    getEnv("REDIS_HOST", "localhost"),
		RedisPort:    redisPort,
		JWTSecret:    getEnv("JWT_SECRET", "super_secret_dev_key_change_in_production_32char_minimum"),
		AIServiceURL: getEnv("AI_SERVICE_URL", "http://localhost:8000"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}
