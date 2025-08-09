package repository_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/repository"
)

var testDB *gorm.DB
var accountRepo repository.AccountRepository

// TestMain sets up the test database and runs all tests.
func TestMain(m *testing.M) {
	// Load environment variables from .env file for tests
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or failed to load .env file for tests. Assuming environment variables are set.")
	}

	// Get test database URL from environment variable
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	log.Printf("Attempting to connect to TEST_DATABASE_URL: %s\n", testDBURL) // ADD THIS LINE

	if testDBURL == "" {
		log.Fatal("TEST_DATABASE_URL environment variable not set. Please set it to your PostgreSQL test database.")
	}

	// Initialize PostgreSQL database for testing
	db, err := gorm.Open(postgres.Open(testDBURL), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				LogLevel: logger.Silent, // Keep logs silent for tests unless debugging
			},
		),
	})
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL test database: %v", err)
	}

	testDB = db
	accountRepo = repository.NewAccountRepository(testDB)

	// Migrate the schema for the test database
	err = testDB.AutoMigrate(&models.Account{})
	if err != nil {
		log.Fatalf("Failed to auto migrate schema for test database: %v", err)
	}

	// Run the tests
	exitCode := m.Run()

	// Clean up (optional for in-memory, but good practice)
	// In a real PostgreSQL test setup, you might drop the test schema/database here if it was created dynamically.
	// sqlDB, _ := testDB.DB()
	// sqlDB.Close()

	os.Exit(exitCode)
}

// clearAccountsTable clears all data from the accounts table before each test.
// Using TRUNCATE for PostgreSQL for speed and to reset auto-incrementing IDs if any.
func clearAccountsTable(t *testing.T) {
	// CASCADE is often needed if there are foreign key constraints, which we don't have yet but will.
	err := testDB.Exec("TRUNCATE TABLE accounts CASCADE").Error
	require.NoError(t, err, "Failed to clear accounts table")
}

func TestCreateAccount(t *testing.T) {
	clearAccountsTable(t)
	ctx := context.Background()

	account := &models.Account{
		Name:     "Test User Wallet",
		Type:     models.Asset,
		UserID:   uuid.New().String(),
		Currency: "USD",
	}

	err := accountRepo.CreateAccount(ctx, account)
	require.NoError(t, err)
	assert.NotEmpty(t, account.ID)
	assert.False(t, account.CreatedAt.IsZero())
	assert.False(t, account.UpdatedAt.IsZero())

	// Verify the account exists in the database
	retrievedAccount, err := accountRepo.GetAccountByID(ctx, account.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedAccount)
	assert.Equal(t, account.ID, retrievedAccount.ID)
	assert.Equal(t, account.Name, retrievedAccount.Name)
	assert.Equal(t, account.Type, retrievedAccount.Type)
	assert.Equal(t, account.UserID, retrievedAccount.UserID)
	assert.Equal(t, account.Currency, retrievedAccount.Currency)
}

func TestGetAccountByID(t *testing.T) {
	clearAccountsTable(t)
	ctx := context.Background()

	// Create an account first
	account := &models.Account{
		Name:     "Another Wallet",
		Type:     models.Liability,
		UserID:   uuid.New().String(),
		Currency: "EUR",
	}
	err := accountRepo.CreateAccount(ctx, account)
	require.NoError(t, err)

	// Test successful retrieval
	retrievedAccount, err := accountRepo.GetAccountByID(ctx, account.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedAccount)
	assert.Equal(t, account.ID, retrievedAccount.ID)

	// Test not found
	nonExistentID := uuid.New().String()
	notFoundAccount, err := accountRepo.GetAccountByID(ctx, nonExistentID)
	require.NoError(t, err)
	assert.Nil(t, notFoundAccount)
}

func TestGetAccountsByUserID(t *testing.T) {
	clearAccountsTable(t)
	ctx := context.Background()

	userID := uuid.New().String()

	// Create multiple accounts for the same user
	account1 := &models.Account{Name: "Wallet A", Type: models.Asset, UserID: userID, Currency: "USD"}
	account2 := &models.Account{Name: "Wallet B", Type: models.Asset, UserID: userID, Currency: "EUR"}
	account3 := &models.Account{Name: "Another User Wallet", Type: models.Asset, UserID: uuid.New().String(), Currency: "GBP"}

	require.NoError(t, accountRepo.CreateAccount(ctx, account1))
	require.NoError(t, accountRepo.CreateAccount(ctx, account2))
	require.NoError(t, accountRepo.CreateAccount(ctx, account3))

	accounts, err := accountRepo.GetAccountsByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Len(t, accounts, 2)

	// Verify the returned accounts belong to the correct user
	foundIDs := make(map[string]bool)
	for _, acc := range accounts {
		assert.Equal(t, userID, acc.UserID)
		foundIDs[acc.ID] = true
	}
	assert.True(t, foundIDs[account1.ID])
	assert.True(t, foundIDs[account2.ID])

	// Test for a user with no accounts
	noAccountsUser := uuid.New().String()
	emptyAccounts, err := accountRepo.GetAccountsByUserID(ctx, noAccountsUser)
	require.NoError(t, err)
	assert.Len(t, emptyAccounts, 0)
}

func TestUpdateAccount(t *testing.T) {
	clearAccountsTable(t)
	ctx := context.Background()

	account := &models.Account{
		Name:     "Original Name",
		Type:     models.Asset,
		UserID:   uuid.New().String(),
		Currency: "JPY",
	}
	require.NoError(t, accountRepo.CreateAccount(ctx, account))

	account.Name = "Updated Name"
	// GORM automatically updates UpdatedAt on .Save() or specific .Update() calls
	err := accountRepo.UpdateAccount(ctx, account)
	require.NoError(t, err)

	retrievedAccount, err := accountRepo.GetAccountByID(ctx, account.ID)
	require.NoError(t, err)
	require.NotNil(t, retrievedAccount)
	assert.Equal(t, "Updated Name", retrievedAccount.Name)
	assert.True(t, retrievedAccount.UpdatedAt.After(account.CreatedAt), "UpdatedAt should be after CreatedAt")

	// Test update of non-existent account
	nonExistentAccount := &models.Account{ID: uuid.New().String(), Name: "Should Not Update"}
	err = accountRepo.UpdateAccount(ctx, nonExistentAccount)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found for update")
}

func TestDeleteAccount(t *testing.T) {
	clearAccountsTable(t)
	ctx := context.Background()

	account := &models.Account{
		Name:     "Account to Delete",
		Type:     models.Expense,
		UserID:   uuid.New().String(),
		Currency: "AUD",
	}
	require.NoError(t, accountRepo.CreateAccount(ctx, account))

	// Verify it exists before deletion
	retrievedAccount, err := accountRepo.GetAccountByID(ctx, account.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrievedAccount)

	// Perform deletion
	err = accountRepo.DeleteAccount(ctx, account.ID)
	require.NoError(t, err)

	// Verify it no longer exists
	deletedAccount, err := accountRepo.GetAccountByID(ctx, account.ID)
	require.NoError(t, err)
	assert.Nil(t, deletedAccount, "Account should be deleted")

	// Test deleting non-existent account
	err = accountRepo.DeleteAccount(ctx, uuid.New().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found for deletion")
}
