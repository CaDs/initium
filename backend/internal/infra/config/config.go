package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	AppEnv        string
	AppURL        string
	HTTPPort      int
	DevBypassAuth bool

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	JWTPrivateKeyPath string
	JWTPublicKeyPath  string

	SMTPHost      string
	SMTPPort      int
	SMTPFrom      string
	AppDeepScheme string
}

// Load reads configuration from environment variables and validates required fields.
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:             envOrDefault("APP_ENV", "development"),
		AppURL:             envOrDefault("APP_URL", "http://localhost:3000"),
		HTTPPort:           envIntOrDefault("HTTP_PORT", 8000),
		DevBypassAuth:      envBoolOrDefault("DEV_BYPASS_AUTH", false),
		DBHost:             envOrDefault("DB_HOST", "localhost"),
		DBPort:             envIntOrDefault("DB_PORT", 5432),
		DBUser:             envOrDefault("DB_USER", "initium"),
		DBPassword:         envOrDefault("DB_PASSWORD", "initium"),
		DBName:             envOrDefault("DB_NAME", "initium_dev"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  envOrDefault("GOOGLE_REDIRECT_URL", "http://localhost:8000/api/auth/google/callback"),
		JWTPrivateKeyPath:  envOrDefault("JWT_PRIVATE_KEY_PATH", "jwt_private.pem"),
		JWTPublicKeyPath:   envOrDefault("JWT_PUBLIC_KEY_PATH", "jwt_public.pem"),
		SMTPHost:           envOrDefault("SMTP_HOST", "localhost"),
		SMTPPort:           envIntOrDefault("SMTP_PORT", 1025),
		SMTPFrom:           envOrDefault("SMTP_FROM", "noreply@initium.local"),
		AppDeepScheme:      envOrDefault("APP_DEEP_SCHEME", "initium"),
	}

	if cfg.DevBypassAuth && cfg.AppEnv != "development" {
		return nil, fmt.Errorf("DEV_BYPASS_AUTH=true is only allowed when APP_ENV=development (got %q)", cfg.AppEnv)
	}

	return cfg, nil
}

// DatabaseDSN returns the PostgreSQL connection string.
func (c *Config) DatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

// DatabaseURL returns the postgres:// URL for golang-migrate.
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envIntOrDefault(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultVal
}

func envBoolOrDefault(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}
