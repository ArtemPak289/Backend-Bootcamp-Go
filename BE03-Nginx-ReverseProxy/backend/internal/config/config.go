package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	InstanceName string
	ReadOnly     bool
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

	instanceName := getEnv("INSTANCE_NAME", "backend-"+port)
	readOnly := getEnvAsBool("READ_ONLY", false)

	return &Config{
		Port:         port,
		DBHost:       dbHost,
		DBPort:       dbPort,
		DBUser:       dbUser,
		DBPassword:   dbPassword,
		DBName:       dbName,
		DBSSLMode:    dbSSLMode,
		InstanceName: instanceName,
		ReadOnly:     readOnly,
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

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val == "true" || val == "1" || val == "yes"
	}
	return defaultVal
}
