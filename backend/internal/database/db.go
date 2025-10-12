package database

import (
	"database/sql"
	"fmt"

	"github.com/raufhm/fairflow/pkg/logger"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.uber.org/zap"
)

// InitDB initializes the PostgreSQL database connection with Bun ORM
// Note: Migrations are NOT run automatically. Run them separately using:
//   go run ./cmd/migrate up
func InitDB(databaseURL string) (*bun.DB, error) {
	// Open PostgreSQL connection
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(databaseURL)))

	// Create Bun DB instance with PostgreSQL dialect
	db := bun.NewDB(sqldb, pgdialect.New())

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	logger.Log.Info("Database connection established", zap.String("note", "Migrations must be run separately with 'go run ./cmd/migrate up'"))

	return db, nil
}
