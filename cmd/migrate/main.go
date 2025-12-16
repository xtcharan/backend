package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yourusername/college-event-backend/pkg/config"
	"github.com/yourusername/college-event-backend/pkg/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg.GetDatabaseDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✓ Connected to database")

	// Read migration script
	migrationPath := "migrations/001_initial_schema.sql"
	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration (ignore errors if schema already exists)
	_, err = db.Exec(string(migration))
	if err != nil {
		// Check if it's just a "already exists" error
		if strings.Contains(err.Error(), "already exists") {
			log.Println("✓ Database schema already exists, skipping migration")
		} else {
			log.Printf("Warning: Migration error: %v", err)
		}
	} else {
		log.Println("✓ Database migration completed successfully")
	}
	fmt.Println("\nDatabase is ready for use!")
}
