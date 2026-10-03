package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Upload   UploadConfig
	CORS     CORSConfig
}
type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	MaxConns int32
	MinConns int32
}
type UploadConfig struct {
	MaxFileSize    int64
	UploadProvider string
}

type CORSConfig struct {
	AllowedOrigin string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	maxConns, _ := strconv.ParseInt(getEnv("DB_MAX_CONNS", "25"), 10, 32)
	minConns, _ := strconv.ParseInt(getEnv("DB_MIN_CONNS", "5"), 10, 32)
	maxUploadSize, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_SIZE", "10485760"), 10, 64)
	config := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "admin"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "costaBackend"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			MaxConns: int32(maxConns),
			MinConns: int32(minConns),
		}, Upload: UploadConfig{
			MaxFileSize:    maxUploadSize,
			UploadProvider: getEnv("UPLOAD_PROVIDER", "s3"),
		},
		CORS: CORSConfig{
			AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:3000"),
		},
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

// DSN returns the key=value format used by pgx/pgxpool.
func (db *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.Name, db.SSLMode,
	)
}

// MigrationDSN returns the postgres:// URL format required by golang-migrate.
// golang-migrate and pgx use different connection string formats.
func (db *DatabaseConfig) MigrationDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, db.SSLMode,
	)
}

func (config *Config) Validate() error {
	if config.Database.Host == "" {
		return fmt.Errorf("config: DB_host is required")
	}
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return key
	}
	return defaultValue
}
