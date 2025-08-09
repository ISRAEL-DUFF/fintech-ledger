package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/api"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/api/handlers"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/db"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/repository"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/service"
	"github.com/joho/godotenv"
)

const (
	appName    = "fintech-ledger"
	appVersion = "1.0.0"
	defaultPort = "8080"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or failed to load .env file. Using environment variables.")
	}

	// Initialize database
	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	// Initialize repositories
	entryRepo := repository.NewEntryRepository(dbConn)
	accountRepo := repository.NewAccountRepository(dbConn)

	// Initialize services
	transactionService := service.NewTransactionService(entryRepo, accountRepo)

	// Initialize API server
	server := api.NewServer()

	// Set up routes
	setupRoutes(server, transactionService)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting server on :%s", port)
		if err := server.Serve(port); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline to wait for (commented out as not currently used)
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	log.Println("Server exiting")
}

// setupRoutes configures all the routes for the application
func setupRoutes(server *api.Server, transactionService *service.TransactionService) {
	// Initialize transaction handler
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	// Mount API routes
	server.MountHandlers(
		// Health check routes
		func(r chi.Router) {
			r.Get("/health", handlers.HealthCheckHandler(appVersion))
			r.Get("/api/v1/health", handlers.HealthCheckHandler(appVersion))
		},
		// Transaction routes
		transactionHandler.RegisterRoutes,
	)
}

// Example usage of the transaction service
// This is a placeholder for demonstration purposes
// Uncomment and modify as needed for testing
/*
func exampleTransaction(transactionService *service.TransactionService) {
	ctx := context.Background()

	// Example transaction data
	entry := &models.Entry{
		Description:     "Sample transaction",
		TransactionType: "transfer",
		ReferenceID:     "ref-123",
		Date:            time.Now(),
		Lines: []models.EntryLine{
			{
				AccountID: "source-account-id",
				Debit:     100.0,  // Money leaving source account
				Credit:    0.0,
			},
			{
				AccountID: "destination-account-id",
				Debit:     0.0,
				Credit:    100.0,  // Money entering destination account
			},
		},
	}

	// Create the transaction
	err := transactionService.CreateEntry(ctx, entry)
	if err != nil {
		log.Printf("Failed to create transaction: %v", err)
		return
	}

	log.Printf("Created transaction with ID: %s", entry.ID)

	// Example: Get transaction by ID
	retrievedEntry, err := transactionService.GetEntryByID(ctx, entry.ID)
	if err != nil {
		log.Printf("Failed to retrieve transaction: %v", err)
		return
	}

	log.Printf("Retrieved transaction: %+v", retrievedEntry)
}
*/
