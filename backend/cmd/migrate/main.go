package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/raufhm/rra/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Get command from arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Execute command
	switch command {
	case "up":
		migrateUp(cfg)
	case "down":
		migrateDown(cfg)
	case "status":
		showStatus(cfg)
	case "create":
		createMigration(os.Args[2:])
	case "reset":
		resetMigrations(cfg)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("FairFlow Migration Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd/migrate <command>")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  up              Apply all pending migrations")
	fmt.Println("  down            Rollback the last migration")
	fmt.Println("  status          Show migration status")
	fmt.Println("  create <name>   Create new migration files")
	fmt.Println("  reset           Reset migration state (DANGEROUS - dev only)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run ./cmd/migrate up")
	fmt.Println("  go run ./cmd/migrate down")
	fmt.Println("  go run ./cmd/migrate status")
	fmt.Println("  go run ./cmd/migrate create add_user_fields")
	fmt.Println("  go run ./cmd/migrate reset")
}

func getMigrate(cfg *config.Config) (*migrate.Migrate, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return m, nil
}

func migrateUp(cfg *config.Config) {
	fmt.Printf("📦 Running migrations for %s environment...\n", cfg.Environment)

	m, err := getMigrate(cfg)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("✅ No new migrations to apply - database is up to date")
			return
		}
		log.Fatalf("❌ Migration failed: %v", err)
	}

	fmt.Println("✅ All migrations applied successfully")
}

func migrateDown(cfg *config.Config) {
	// Safety check for production
	if cfg.Environment == "production" {
		fmt.Println("⚠️  WARNING: You are about to rollback a migration in PRODUCTION")
		fmt.Print("Type 'yes' to continue: ")
		var response string
		fmt.Scanln(&response)
		if response != "yes" {
			fmt.Println("❌ Rollback cancelled")
			return
		}
	}

	fmt.Printf("🔄 Rolling back last migration in %s...\n", cfg.Environment)

	m, err := getMigrate(cfg)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	err = m.Steps(-1)
	if err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("ℹ️  No migrations to roll back")
			return
		}
		log.Fatalf("❌ Rollback failed: %v", err)
	}

	fmt.Println("✅ Migration rolled back successfully")
}

func showStatus(cfg *config.Config) {
	fmt.Printf("📊 Migration status for %s environment:\n\n", cfg.Environment)

	m, err := getMigrate(cfg)
	if err != nil {
		log.Fatalf("❌ %v", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			fmt.Println("  Status: No migrations applied yet")
			fmt.Println("  Next: Run 'go run ./cmd/migrate up' to apply migrations")
			return
		}
		log.Fatalf("❌ Failed to get migration status: %v", err)
	}

	fmt.Printf("  Current Version: %d\n", version)
	if dirty {
		fmt.Println("  Status: ⚠️  DIRTY - Migration failed, needs manual fix")
		fmt.Println("  Action: Fix the issue and run 'go run ./cmd/migrate reset' if needed")
	} else {
		fmt.Println("  Status: ✅ Clean")
		fmt.Println("  Next: Create new migrations or run 'go run ./cmd/migrate up'")
	}
}

func createMigration(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Please provide a migration name")
		fmt.Println("Usage: go run ./cmd/migrate create <name>")
		fmt.Println("Example: go run ./cmd/migrate create add_user_fields")
		os.Exit(1)
	}

	name := args[0]
	timestamp := time.Now().Unix()

	upFile := fmt.Sprintf("migrations/%d_%s.up.sql", timestamp, name)
	downFile := fmt.Sprintf("migrations/%d_%s.down.sql", timestamp, name)

	// Create up migration
	upContent := fmt.Sprintf("-- Migration: %s\n-- Created: %s\n\n-- Add your UP migration SQL here\n", name, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		log.Fatalf("❌ Failed to create up migration: %v", err)
	}

	// Create down migration
	downContent := fmt.Sprintf("-- Migration: %s\n-- Created: %s\n\n-- Add your DOWN (rollback) migration SQL here\n", name, time.Now().Format(time.RFC3339))
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		log.Fatalf("❌ Failed to create down migration: %v", err)
	}

	fmt.Println("✅ Migration files created:")
	fmt.Printf("   📄 %s\n", upFile)
	fmt.Printf("   📄 %s\n", downFile)
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit the migration files to add your SQL")
	fmt.Println("  2. Run 'go run ./cmd/migrate up' to apply the migration")
}

func resetMigrations(cfg *config.Config) {
	// Safety check for non-development environments
	if cfg.Environment != "development" {
		fmt.Printf("❌ Migration reset is only allowed in development environment\n")
		fmt.Printf("   Current environment: %s\n", cfg.Environment)
		os.Exit(1)
	}

	fmt.Println("⚠️  WARNING: This will reset the migration state")
	fmt.Println("   All migration history will be lost")
	fmt.Print("Type 'yes' to continue: ")
	var response string
	fmt.Scanln(&response)
	if response != "yes" {
		fmt.Println("❌ Reset cancelled")
		return
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("DROP TABLE IF EXISTS schema_migrations")
	if err != nil {
		log.Fatalf("❌ Failed to drop schema_migrations table: %v", err)
	}

	fmt.Println("✅ Migration state reset successfully")
	fmt.Println("   Run 'go run ./cmd/migrate up' to apply migrations from scratch")
}
