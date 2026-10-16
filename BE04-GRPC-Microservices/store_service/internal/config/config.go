package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DBHost              string
	DBPort              string
	DBUser              string
	DBPassword          string
	DBName              string
	DBSSLMode           string
	AuthServiceGRPCAddr string
}

func Load() *Config {
	_ = godotenv.Load()

	port := getEnv("PORT", "")
	if port == "" {
		port = getEnv("SERVER_PORT", "8080")
	}

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "")
	if dbUser == "" {
		// Fallback to postgres or current system user
		if u := os.Getenv("USER"); u != "" {
			dbUser = u
		} else {
			dbUser = "postgres"
		}
	}

	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "store_db")
	dbSSLMode := getEnv("DB_SSLMODE", "disable")

	authGRPCAddr := getEnv("AUTH_SERVICE_ADDR", "")
	if authGRPCAddr == "" {
		authGRPCAddr = getEnv("AUTH_GRPC_ADDR", "localhost:50051")
	}

	return &Config{
		Port:                port,
		DBHost:              dbHost,
		DBPort:              dbPort,
		DBUser:              dbUser,
		DBPassword:          dbPassword,
		DBName:              dbName,
		DBSSLMode:           dbSSLMode,
		AuthServiceGRPCAddr: authGRPCAddr,
	}
}

func (c *Config) DSN() string {
	if c.DBPassword != "" {
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode)
	}
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBName, c.DBSSLMode)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
