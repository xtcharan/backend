package main

import (
	"fmt"
	"log"

	"github.com/yourusername/college-event-backend/internal/api"
	"github.com/yourusername/college-event-backend/internal/services/auth"
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

	// Setup router
	router := api.NewRouter(db, authService, cfg.CORSAllowedOrigins)
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
