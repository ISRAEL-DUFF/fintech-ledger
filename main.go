package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/db"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/repository"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or failed to load .env file. Assuming environment variables are set.")
	}

	// Set up database connection using GORM
	dbConn, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// AutoMigrate the schema. This will create/update tables based on the models.
	// In a production environment, prefer dedicated migration tools like golang-migrate for controlled rollouts.
	err = dbConn.AutoMigrate(&models.Account{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database schema: %v", err)
	}
	log.Println("Database schema auto-migrated successfully.")

	// Initialize Account Repository
	accountRepo := repository.NewAccountRepository(dbConn)

	ctx := context.Background()

	// Demonstrate CreateAccount
	newAccount := &models.Account{
		Name:      "Jane Doe's Wallet",
		Type:      models.Asset,
		UserID:    "user-456",
		Currency:  "EUR",
	}

	log.Printf("Attempting to create account: %+v\n", newAccount)

	err = accountRepo.CreateAccount(ctx, newAccount)
	if err != nil {
		log.Fatalf("Failed to create account: %v", err)
	}

	log.Printf("Account created successfully with ID: %s\n", newAccount.ID)

	// Demonstrate GetAccountByID
	retrievedAccount, err := accountRepo.GetAccountByID(ctx, newAccount.ID)
	if err != nil {
		log.Fatalf("Failed to retrieve account: %v", err)
	}
	if retrievedAccount != nil {
		log.Printf("Retrieved account: %+v\n", retrievedAccount)
	} else {
		log.Printf("Account with ID %s not found.\n", newAccount.ID)
	}

	// Demonstrate GetAccountsByUserID
	userAccounts, err := accountRepo.GetAccountsByUserID(ctx, "user-456")
	if err != nil {
		log.Fatalf("Failed to retrieve accounts by user ID: %v", err)
	}
	log.Printf("Accounts for user-456: %+v\n", userAccounts)

	// Demonstrate UpdateAccount
	if retrievedAccount != nil {
		retrievedAccount.Name = "Jane Doe's Euro Wallet"
		err = accountRepo.UpdateAccount(ctx, retrievedAccount)
		if err != nil {
			log.Fatalf("Failed to update account: %v", err)
		}
		log.Printf("Account updated successfully: %+v\n", retrievedAccount)
	}

	// Demonstrate DeleteAccount (use with caution in real systems)
	// err = accountRepo.DeleteAccount(ctx, newAccount.ID)
	// if err != nil {
	// 	log.Fatalf("Failed to delete account: %v", err)
	// }
	// log.Printf("Account with ID %s deleted successfully.\n", newAccount.ID)

}
