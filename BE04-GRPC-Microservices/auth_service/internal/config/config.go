package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GRPCPort         string
	DBHost           string
	DBPort           string
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	DatabaseURL      string
	JWTSecret        string
	TokenExpiryHours int
}

func Load() *Config {
	port := getEnv("GRPC_PORT", "")
	if port == "" {
		port = getEnv("PORT", "50051")
	}

	expiryHours, err := strconv.Atoi(getEnv("TOKEN_EXPIRY_HOURS", "24"))
	if err != nil || expiryHours <= 0 {
		expiryHours = 24
	}

	return &Config{
		GRPCPort:         port,
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnv("DB_PORT", "5432"),
		DBUser:           getEnv("DB_USER", "artempak"),
		DBPassword:       getEnv("DB_PASSWORD", ""),
		DBName:           getEnv("DB_NAME", "auth_db"),
		DBSSLMode:        getEnv("DB_SSLMODE", "disable"),
		DatabaseURL:      getEnv("DATABASE_URL", ""),
		JWTSecret:        getEnv("JWT_SECRET", "school21-secret-jwt-token-key-2026"),
		TokenExpiryHours: expiryHours,
	}
}

func (c *Config) ConnectionString() string {
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	connStr := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
	if c.DBUser != "" {
		connStr += fmt.Sprintf(" user=%s", c.DBUser)
	}
	if c.DBPassword != "" {
		connStr += fmt.Sprintf(" password=%s", c.DBPassword)
	}
	return connStr
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
