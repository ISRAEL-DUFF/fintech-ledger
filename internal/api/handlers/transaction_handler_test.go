package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/api/dto"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Use the TransactionService interface from the service package

// mockTransactionService is a mock implementation of the TransactionService interface
type mockTransactionService struct {
	createEntryFunc func(ctx context.Context, entry *models.Entry) error
}

// Ensure mockTransactionService implements service.TransactionService
var _ service.TransactionService = (*mockTransactionService)(nil)

func (m *mockTransactionService) CreateEntry(ctx context.Context, entry *models.Entry) error {
	if m.createEntryFunc != nil {
		return m.createEntryFunc(ctx, entry)
	}
	return nil
}

func (m *mockTransactionService) GetEntryByID(ctx context.Context, id string) (*models.Entry, error) {
	// Return a sample entry for testing
	return &models.Entry{
		ID:              id,
		Description:     "Test Entry",
		TransactionType: "transfer",
		ReferenceID:     "test-ref-123",
		Date:            time.Now(),
	}, nil
}

func (m *mockTransactionService) GetEntriesByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*models.Entry, int64, error) {
	// Return empty results for now
	return []*models.Entry{}, 0, nil
}

func (m *mockTransactionService) ValidateEntry(ctx context.Context, entry *models.Entry) error {
	// Simple validation for testing
	if entry.Description == "" {
		return fmt.Errorf("description is required")
	}
	if len(entry.Lines) < 2 {
		return fmt.Errorf("at least two entry lines are required")
	}
	// Additional validation for test cases
	if entry.TransactionType == "invalid-type" {
		return fmt.Errorf("invalid transaction type")
	}
	return nil
}

func (m *mockTransactionService) ProcessTransfer(ctx context.Context, req service.TransferRequest) (*models.Transaction, error) {
	return &models.Transaction{
		ID:          "test-transfer-123",
		Type:        "transfer",
		Status:      "completed",
		Description: fmt.Sprintf("Transfer of %f %s", req.Amount, req.Currency),
		CreatedAt:   time.Now(),
	}, nil
}

func (m *mockTransactionService) ProcessDeposit(ctx context.Context, req service.DepositRequest) (*models.Transaction, error) {
	return &models.Transaction{
		ID:          "test-deposit-123",
		Type:        "deposit",
		Status:      "completed",
		Description: fmt.Sprintf("Deposit of %f %s", req.Amount, req.Currency),
		CreatedAt:   time.Now(),
	}, nil
}

func (m *mockTransactionService) ProcessWithdrawal(ctx context.Context, req service.WithdrawalRequest) (*models.Transaction, error) {
	return &models.Transaction{
		ID:          "test-withdrawal-123",
		Type:        "withdrawal",
		Status:      "completed",
		Description: fmt.Sprintf("Withdrawal of %f %s", req.Amount, req.Currency),
		CreatedAt:   time.Now(),
	}, nil
}

func (m *mockTransactionService) ProcessExchange(ctx context.Context, req service.ExchangeRequest) (*models.Transaction, error) {
	return &models.Transaction{
		ID:          "test-exchange-123",
		Type:        "exchange",
		Status:      "completed",
		Description: fmt.Sprintf("Exchange %f %s to %f %s", 
			req.SourceAmount, req.SourceCurrency, 
			req.DestinationAmount, req.DestinationCurrency),
		CreatedAt:   time.Now(),
	}, nil
}

func (m *mockTransactionService) ProcessFee(ctx context.Context, req service.FeeRequest) (*models.Transaction, error) {
	return &models.Transaction{
		ID:          "test-fee-123",
		Type:        "fee",
		Status:      "completed",
		Description: fmt.Sprintf("Fee of %f %s", req.Amount, req.Currency),
		CreatedAt:   time.Now(),
	}, nil
}

func (m *mockTransactionService) ReverseTransfer(ctx context.Context, transactionID string) error {
	return nil
}

func (m *mockTransactionService) ReverseDeposit(ctx context.Context, transactionID string) error {
	return nil
}

func (m *mockTransactionService) ReverseWithdrawal(ctx context.Context, transactionID string) error {
	return nil
}

func (m *mockTransactionService) ReverseExchange(ctx context.Context, transactionID string) error {
	return nil
}

func (m *mockTransactionService) ReverseFee(ctx context.Context, transactionID string) error {
	return nil
}

func TestTransactionHandler_CreateTransaction(t *testing.T) {
	now := time.Now()
	// Setup test cases
	tests := []struct {
		name           string
		request        dto.CreateTransactionRequest
		setupMocks     func(*mockTransactionService)
		expectedStatus int
		expectedBody   map[string]interface{}
		shouldPanic    bool
	}{
		// Happy path
		{
			name: "successful transaction creation",
			request: dto.CreateTransactionRequest{
				Description:     "Test transaction",
				TransactionType: "transfer",
				ReferenceID:     "test-ref-123",
				Date:            now,
				Lines: []dto.TransactionLineEntry{
					{AccountID: "550e8400-e29b-41d4-a716-446655440000", Amount: -100.00, Currency: "USD"},
					{AccountID: "550e8400-e29b-41d4-a716-446655440001", Amount: 100.00, Currency: "USD"},
				},
			},
			setupMocks: func(mockSvc *mockTransactionService) {
				mockSvc.createEntryFunc = func(ctx context.Context, entry *models.Entry) error {
					// Set a dummy ID on the entry to simulate database creation
					entry.ID = "550e8400-e29b-41d4-a716-446655440002"
					entry.CreatedAt = now
					entry.UpdatedAt = now
					return nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"description":      "Test transaction",
				"transaction_type": "transfer",
			},
		},
		// Error cases
		{
			name: "validation error - missing required fields",
			request: dto.CreateTransactionRequest{
				Description: "", // Missing required field
				Lines: []dto.TransactionLineEntry{
					{AccountID: "550e8400-e29b-41d4-a716-446655440000", Amount: -100.00, Currency: "USD"},
				},
			},
			setupMocks:     nil, // No mocks needed for validation failure
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{},
		},
		{
			name: "service error - invalid transaction",
			request: dto.CreateTransactionRequest{
				Description:     "Invalid transaction",
				TransactionType: "invalid-type",
				ReferenceID:     "test-ref-124",
				Date:            now,
				Lines: []dto.TransactionLineEntry{
					{AccountID: "550e8400-e29b-41d4-a716-446655440000", Amount: -100.00, Currency: "USD"},
					{AccountID: "550e8400-e29b-41d4-a716-446655440001", Amount: 100.00, Currency: "USD"},
				},
			},
			setupMocks: func(mockSvc *mockTransactionService) {
				mockSvc.createEntryFunc = func(ctx context.Context, entry *models.Entry) error {
					return fmt.Errorf("invalid transaction type")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   map[string]interface{}{},
		},
		{
			name: "invalid request body",
			request: dto.CreateTransactionRequest{
				Description: "Test transaction",
				Lines: []dto.TransactionLineEntry{
					{AccountID: "", Amount: 0, Currency: ""}, // Invalid line data
				},
			},
			setupMocks:     nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock service
			mockSvc := &mockTransactionService{}

			// Create handler with mock service
			handler := &TransactionHandler{
				transactionService: mockSvc,
			}
			
			// Set up the mock behavior
			if tt.setupMocks != nil {
				tt.setupMocks(mockSvc)
			}

			// Skip test if it's marked as should panic
			if tt.shouldPanic {
				t.Skip("Skipping test that is expected to panic")
			}

			// Create request
			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest("POST", "/api/v1/transactions", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create router with validation middleware
			r := chi.NewRouter()
			// Create a validation middleware that validates the request data
			validationMiddleware := func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// For test purposes, we'll validate the request data here
					var req dto.CreateTransactionRequest
					if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
						http.Error(w, "Invalid request body", http.StatusBadRequest)
						return
					}

					// Validate required fields
					if req.Description == "" || len(req.Lines) == 0 {
						http.Error(w, "Missing required fields", http.StatusBadRequest)
						return
					}

					// Validate transaction lines
					for _, line := range req.Lines {
						if line.AccountID == "" || line.Currency == "" {
							http.Error(w, "Invalid transaction line", http.StatusBadRequest)
							return
						}
					}

					// Store the validated data in the context
					ctx := context.WithValue(r.Context(), "validatedData", &req)
					next.ServeHTTP(w, r.WithContext(ctx))
				})
			}

			// Apply the validation middleware
			r.With(validationMiddleware).Post("/api/v1/transactions", handler.CreateTransaction)
			r.ServeHTTP(rr, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Parse response body if we expect a JSON response
			if tt.expectedStatus >= 200 && tt.expectedStatus < 300 && tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err, "Failed to parse response body")

				// Assert expected fields in response
				for key, expectedValue := range tt.expectedBody {
					assert.Equal(t, expectedValue, response[key], "Mismatch in field: %s", key)
				}
			} else if tt.expectedStatus >= 400 {
				// For error responses, just verify the status code
				// The actual error message might be in plain text
				assert.NotEmpty(t, rr.Body.String(), "Expected an error message")
			}

			// No need to verify expectations with our simple mock
		})
	}
}

// NewTestTransactionHandler creates a new TransactionHandler with a mock service
func NewTestTransactionHandler() (*TransactionHandler, *mockTransactionService) {
	mockSvc := &mockTransactionService{}
	handler := &TransactionHandler{
		transactionService: mockSvc,
	}
	return handler, mockSvc
}
