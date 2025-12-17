package main

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
	"github.com/yourusername/college-event-backend/internal/api"
	"github.com/yourusername/college-event-backend/internal/services/auth"
	localstorage "github.com/yourusername/college-event-backend/internal/storage"
	"github.com/yourusername/college-event-backend/pkg/config"
	"github.com/yourusername/college-event-backend/pkg/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting College Event Management API in %s mode...", cfg.Env)

	// Connect to database
	db, err := database.Connect(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Connected to database")

	// Initialize auth service
	authService := auth.NewService(cfg.JWTSecret, cfg.JWTExpiryHours, cfg.RefreshTokenExpiryDays)

	// Initialize storage service based on configuration
	storageService, err := initStorageService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage service: %v", err)
	}
	log.Printf("✓ Storage service initialized (provider: %s)", cfg.StorageProvider)

	// Setup router
	router := api.NewRouter(db, authService, storageService, cfg.CORSAllowedOrigins)
	router.Setup()

	log.Println("✓ API routes configured")
	log.Println("✓ Middleware initialized")

	// Create initial admin user if configured
	if cfg.InitialAdminEmail != "" && cfg.InitialAdminPassword != "" {
		createInitialAdmin(db, authService, cfg)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Server running on http://localhost%s", addr)
	log.Println("API Documentation: http://localhost" + addr + "/health")

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initStorageService creates the appropriate storage service based on configuration
func initStorageService(cfg *config.Config) (localstorage.StorageService, error) {
	switch cfg.StorageProvider {
	case "gcs":
		// Initialize GCS client
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCS client: %w", err)
		}
		log.Printf("  → GCS bucket: %s", cfg.GCSBucketName)
		if cfg.GCSCdnURL != "" {
			log.Printf("  → CDN URL: %s", cfg.GCSCdnURL)
		}
		return localstorage.NewGCSStorage(client, cfg.GCSBucketName, cfg.GCSCdnURL), nil

	case "local":
		fallthrough
	default:
		// Use local storage for development
		basePath := "./uploads"
		baseURL := fmt.Sprintf("http://localhost:%s/uploads", cfg.Port)
		log.Printf("  → Local storage: %s", basePath)
		return localstorage.NewLocalStorage(basePath, baseURL), nil
	}
}

func createInitialAdmin(db *database.DB, authService *auth.Service, cfg *config.Config) {
	// Check if admin already exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", cfg.InitialAdminEmail).Scan(&exists)
	if err != nil || exists {
		return
	}

	// Hash password
	passwordHash, err := authService.HashPassword(cfg.InitialAdminPassword)
	if err != nil {
		log.Printf("Warning: Failed to create initial admin: %v", err)
		return
	}

	// Create admin user
	_, err = db.Exec(`
		INSERT INTO users (email, password_hash, full_name, role)
		VALUES ($1, $2, $3, $4)
	`, cfg.InitialAdminEmail, passwordHash, "Initial Admin", "admin")

	if err != nil {
		log.Printf("Warning: Failed to create initial admin: %v", err)
	} else {
		log.Printf("✓ Created initial admin user: %s", cfg.InitialAdminEmail)
	}
}
