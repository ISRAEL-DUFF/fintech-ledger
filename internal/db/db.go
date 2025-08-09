package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes and returns a new GORM database connection configured for PostgreSQL.
func InitDB() (*gorm.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:              time.Second,   // Slow SQL threshold
				LogLevel:                   logger.Info, // Log level
				IgnoreRecordNotFoundError: true,            // Ignore ErrRecordNotFound error for logger
				Colorful:                   true,            // Disable color
			},
		),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open database connection with GORM (PostgreSQL): %w", err)
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings (example values)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Successfully connected to the database with GORM (PostgreSQL)!")
	return db, nil
}

// InitTestDB initializes and returns a new GORM database connection for testing.
// It uses SQLite by default for testing purposes.
func InitTestDB(dsn string) (*gorm.DB, error) {
	// Default to in-memory SQLite if no DSN is provided
	if dsn == "" {
		dsn = "file::memory:?cache=shared"
	}

	// Use SQLite for testing
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // Disable logging during tests
	})

	if err != nil {
		return nil, fmt.Errorf("failed to open test database connection: %w", err)
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings for testing
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
