package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string
	Env  string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis
	RedisHost     string
	RedisPort     string
	RedisPassword string

	// JWT
	JWTSecret              string
	JWTExpiryHours         int
	RefreshTokenExpiryDays int

	// Storage
	StorageProvider string
	GCSBucketName   string
	GCSProjectID    string
	AWSRegion       string
	AWSBucketName   string
	AWSAccessKeyID  string
	AWSSecretKey    string

	// CORS
	CORSAllowedOrigins string

	// Rate Limiting
	RateLimitRequestsPerMinute int

	// Admin
	InitialAdminEmail    string
	InitialAdminPassword string
}

func Load() (*Config, error) {
	// Load .env file if exists (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Port:                       getEnv("PORT", "8080"),
		Env:                        getEnv("ENV", "development"),
		DBHost:                     getEnv("DB_HOST", "localhost"),
		DBPort:                     getEnv("DB_PORT", "5432"),
		DBUser:                     getEnv("DB_USER", "postgres"),
		DBPassword:                 getEnv("DB_PASSWORD", ""),
		DBName:                     getEnv("DB_NAME", "college_events"),
		DBSSLMode:                  getEnv("DB_SSL_MODE", "disable"),
		RedisHost:                  getEnv("REDIS_HOST", "localhost"),
		RedisPort:                  getEnv("REDIS_PORT", "6379"),
		RedisPassword:              getEnv("REDIS_PASSWORD", ""),
		JWTSecret:                  getEnv("JWT_SECRET", ""),
		JWTExpiryHours:             getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		RefreshTokenExpiryDays:     getEnvAsInt("REFRESH_TOKEN_EXPIRY_DAYS", 30),
		StorageProvider:            getEnv("STORAGE_PROVIDER", "local"),
		GCSBucketName:              getEnv("GCS_BUCKET_NAME", ""),
		GCSProjectID:               getEnv("GCS_PROJECT_ID", ""),
		AWSRegion:                  getEnv("AWS_REGION", "us-east-1"),
		AWSBucketName:              getEnv("AWS_BUCKET_NAME", ""),
		AWSAccessKeyID:             getEnv("AWS_ACCESS_KEY_ID", ""),
		AWSSecretKey:               getEnv("AWS_SECRET_ACCESS_KEY", ""),
		CORSAllowedOrigins:         getEnv("CORS_ALLOWED_ORIGINS", "*"),
		RateLimitRequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
		InitialAdminEmail:          getEnv("INITIAL_ADMIN_EMAIL", "admin@college.edu"),
		InitialAdminPassword:       getEnv("INITIAL_ADMIN_PASSWORD", ""),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if c.DBPassword == "" && c.Env == "production" {
		return fmt.Errorf("DB_PASSWORD is required in production")
	}
	return nil
}

func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
