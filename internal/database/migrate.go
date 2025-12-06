package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations applies all pending SQL migrations from the migrations directory
func RunMigrations() error {
	// Create migrations tracking table if it doesn't exist
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := DB.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of applied migrations
	appliedMigrations, err := getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Read migration files
	migrationsDir := "migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		// If migrations directory doesn't exist, skip migrations
		log.Println("⚠️  No migrations directory found, skipping migrations")
		return nil
	}

	// Sort migration files
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// Apply pending migrations
	pendingCount := 0
	for _, filename := range migrationFiles {
		version := strings.TrimSuffix(filename, ".sql")

		// Skip if already applied
		if _, exists := appliedMigrations[version]; exists {
			continue
		}

		// Read migration file
		filePath := filepath.Join(migrationsDir, filename)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", filename, err)
		}

		// Execute migration
		log.Printf("📦 Applying migration: %s", filename)
		if _, err := DB.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", filename, err)
		}

		// Record migration as applied
		if err := recordMigration(version); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", filename, err)
		}

		pendingCount++
		log.Printf("✓ Migration applied: %s", filename)
	}

	if pendingCount == 0 {
		log.Println("✓ All migrations up to date")
	} else {
		log.Printf("✓ Applied %d migration(s)", pendingCount)
	}

	return nil
}

func getAppliedMigrations() (map[string]bool, error) {
	rows, err := DB.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func recordMigration(version string) error {
	_, err := DB.Exec(
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		version,
	)
	return err
}
