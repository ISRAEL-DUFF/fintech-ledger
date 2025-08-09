package db

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes and returns a new GORM database connection.
func InitDB() (*gorm.DB, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	var dialector gorm.Dialector
	var db *gorm.DB
	var err error

	if strings.HasPrefix(dbURL, "postgres") {
		log.Println("Connecting to PostgreSQL...")
		dialector = postgres.Open(dbURL)
	} else if strings.HasPrefix(dbURL, "file:") || !strings.Contains(dbURL, "://") {
		// Assuming file: for explicit path or no scheme for relative path (e.g., "./test.db")
		log.Println("Connecting to SQLite...")
		dialector = sqlite.Open(dbURL)
	} else {
		return nil, fmt.Errorf("unsupported database URL scheme: %s", dbURL)
	}

	db, err = gorm.Open(dialector, &gorm.Config{
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
		return nil, fmt.Errorf("failed to open database connection with GORM: %w", err)
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

	log.Println("Successfully connected to the database with GORM!")
	return db, nil
}
