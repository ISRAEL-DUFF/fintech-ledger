package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/api/dto"
	handlers "github.com/ISRAEL-DUFF/fintech-ledger/internal/api/handlers"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/db"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/repository"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type testServer struct {
	server *http.Server
	db     *gorm.DB
}

func setupTestServer(t *testing.T) *testServer {
	// Use a test database
	testDB := setupTestDB(t)

	// Initialize repositories
	entryRepo := repository.NewEntryRepository(testDB)
	accountRepo := repository.NewAccountRepository(testDB)

	// Create test accounts
	account1 := &models.Account{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		Name:      "Test Account 1",
		Type:      "asset",
		Currency:  "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	account2 := &models.Account{
		ID:        "550e8400-e29b-41d4-a716-446655440001",
		Name:      "Test Account 2",
		Type:      "expense",
		Currency:  "USD",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(t, accountRepo.CreateAccount(context.Background(), account1))
	require.NoError(t, accountRepo.CreateAccount(context.Background(), account2))

	// Initialize services
	transactionService := service.NewTransactionService(entryRepo, accountRepo)

	// Initialize handler
	handler := handlers.NewTransactionHandler(transactionService)

	// Create router
	r := chi.NewRouter()
	handler.RegisterRoutes(r)

	// Create test server
	ts := &testServer{
		server: &http.Server{
			Handler: r,
		},
		db: testDB,
	}

	return ts
}

func setupTestDB(t *testing.T) *gorm.DB {
	// Use a test database URL or in-memory SQLite
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		testDBURL = "file:test.db?mode=memory&cache=shared"
	}

	db, err := db.InitTestDB(testDBURL)
	require.NoError(t, err, "Failed to initialize test database")

	// Run migrations
	err = db.AutoMigrate(
		&models.Account{},
		&models.Entry{},
		&models.EntryLine{},
	)
	require.NoError(t, err, "Failed to run migrations")

	return db
}

func (ts *testServer) teardown() {
	// Clean up test database
	sqlDB, err := ts.db.DB()
	if err == nil {
		sqlDB.Close()
	}
}

func TestTransactionHandler_CreateTransaction_Integration(t *testing.T) {
	// Skip if running short tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := setupTestServer(t)
	defer ts.teardown()

	tests := []struct {
		name           string
		request        dto.CreateTransactionRequest
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful transaction creation",
			request: dto.CreateTransactionRequest{
				Description:     "Test transaction",
				TransactionType: "transfer",
				ReferenceID:     "test-ref-123",
				Date:            time.Now(),
				Lines: []dto.TransactionLineEntry{
					{AccountID: "550e8400-e29b-41d4-a716-446655440000", Amount: -100.00, Currency: "USD"},
					{AccountID: "550e8400-e29b-41d4-a716-446655440001", Amount: 100.00, Currency: "USD"},
				},
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"description":      "Test transaction",
				"transaction_type": "transfer",
			},
		},
		{
			name: "validation error - missing required fields",
			request: dto.CreateTransactionRequest{
				// Missing description
				Lines: []dto.TransactionLineEntry{
					{AccountID: "550e8400-e29b-41d4-a716-446655440000", Amount: -100.00, Currency: "USD"},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error - invalid account",
			request: dto.CreateTransactionRequest{
				Description:     "Test transaction",
				TransactionType: "transfer",
				ReferenceID:     "test-ref-124",
				Date:            time.Now(),
				Lines: []dto.TransactionLineEntry{
					{AccountID: "invalid-account-id", Amount: -100.00, Currency: "USD"},
				},
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			body, err := json.Marshal(tt.request)
			require.NoError(t, err, "Failed to marshal request body")

			req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create router and serve the request
			r := chi.NewRouter()
			handler := handlers.NewTransactionHandler(
				service.NewTransactionService(
					repository.NewEntryRepository(ts.db),
					repository.NewAccountRepository(ts.db),
				),
			)
			handler.RegisterRoutes(r)
			r.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "Unexpected status code")

			// Parse response body if we expect a JSON response
			if rr.Code >= 200 && rr.Code < 300 && tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err, "Failed to parse response body")

				// Assert expected fields in response
				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key], "Mismatch in field: %s", key)
				}
			} else if rr.Code >= 400 {
				// For error responses, just verify we got an error message
				assert.NotEmpty(t, rr.Body.String(), "Expected an error message")
			}
		})
	}
}

func TestTransactionHandler_GetTransaction_Integration(t *testing.T) {
	// Skip if running short tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ts := setupTestServer(t)
	defer ts.teardown()

	// Create a test transaction first
	transactionReq := dto.CreateTransactionRequest{
		Description:     "Test Get Transaction",
		TransactionType: "transfer",
		ReferenceID:     "test-get-ref-123",
		Date:            time.Now(),
		Lines: []dto.TransactionLineEntry{
			{AccountID: "550e8400-e29b-41d4-a716-446655440000", Amount: -100.00, Currency: "USD"},
			{AccountID: "550e8400-e29b-41d4-a716-446655440001", Amount: 100.00, Currency: "USD"},
		},
	}

	// Create the transaction
	body, err := json.Marshal(transactionReq)
	require.NoError(t, err, "Failed to marshal request body")

	req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	handler := handlers.NewTransactionHandler(
		service.NewTransactionService(
			repository.NewEntryRepository(ts.db),
			repository.NewAccountRepository(ts.db),
		),
	)
	handler.RegisterRoutes(r)
	r.ServeHTTP(rr, req)

	// Assert the transaction was created
	assert.Equal(t, http.StatusCreated, rr.Code, "Failed to create test transaction")

	// Extract the created transaction ID
	var createdTx map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &createdTx)
	require.NoError(t, err, "Failed to parse created transaction response")

	txID, ok := createdTx["id"].(string)
	require.True(t, ok, "Failed to extract transaction ID from response")

	// Test getting the transaction
	t.Run("get existing transaction", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/transactions/%s", txID), nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Failed to get transaction")

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err, "Failed to parse response body")

		assert.Equal(t, "Test Get Transaction", response["description"])
		assert.Equal(t, "transfer", response["transaction_type"])
	})

	// Test getting non-existent transaction
	t.Run("get non-existent transaction", func(t *testing.T) {
		nonExistentID := "00000000-0000-0000-0000-000000000000"
		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/transactions/%s", nonExistentID), nil)
		rr := httptest.NewRecorder()

		r.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code, "Expected 404 for non-existent transaction")
	})
}
