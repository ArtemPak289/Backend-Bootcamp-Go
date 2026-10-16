package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"auth_service/internal/config"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	connStr := cfg.ConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		// Attempt to auto-create database if it doesn't exist
		if cfg.DatabaseURL == "" {
			_ = ensureDatabaseExists(cfg)
		}
		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}
	}

	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func ensureDatabaseExists(cfg *config.Config) error {
	defaultConn := fmt.Sprintf("host=%s port=%s dbname=postgres sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBSSLMode)
	if cfg.DBUser != "" {
		defaultConn += fmt.Sprintf(" user=%s", cfg.DBUser)
	}
	if cfg.DBPassword != "" {
		defaultConn += fmt.Sprintf(" password=%s", cfg.DBPassword)
	}

	db, err := sql.Open("postgres", defaultConn)
	if err != nil {
		return err
	}
	defer db.Close()

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", cfg.DBName)
	if err := db.QueryRow(query).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		log.Printf("Database %s does not exist, creating...", cfg.DBName)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
		if err != nil {
			return err
		}
		log.Printf("Database %s created successfully", cfg.DBName)
	}
	return nil
}

func RunMigrations(db *sql.DB) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			phone_number VARCHAR(50) NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			salt VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("migration query failed: %w (query: %s)", err, q)
		}
	}
	return nil
}
