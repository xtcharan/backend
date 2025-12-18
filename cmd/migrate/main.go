package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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

	// Read all migration files
	migrationDir := "migrations"
	files, err := os.ReadDir(migrationDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Sort files to ensure order (001, 002, etc.)
	var migrationFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			migrationFiles = append(migrationFiles, f.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Execute each migration
	for _, fileName := range migrationFiles {
		migrationPath := filepath.Join(migrationDir, fileName)
		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", fileName, err)
		}

		// Execute migration (ignore errors if schema already exists)
		_, err = db.Exec(string(migration))
		if err != nil {
			// Check if it's just a "already exists" error
			if strings.Contains(err.Error(), "already exists") ||
				strings.Contains(err.Error(), "duplicate key") {
				log.Printf("✓ %s: Schema already exists, skipping", fileName)
			} else {
				log.Printf("Warning: %s: Migration error: %v", fileName, err)
			}
		} else {
			log.Printf("✓ %s: Migration completed successfully", fileName)
		}
	}

	fmt.Println("\nDatabase is ready for use!")
}
