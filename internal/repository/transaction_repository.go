package repository

import (
	"context"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
)

// TransactionRepository defines the interface for transaction operations
type TransactionRepository interface {
	// CreateTransaction creates a new transaction
	CreateTransaction(ctx context.Context, tx *models.Transaction) error
	
	// GetTransactionByID retrieves a transaction by its ID
	GetTransactionByID(ctx context.Context, id string) (*models.Transaction, error)
	
	// UpdateTransaction updates an existing transaction
	UpdateTransaction(ctx context.Context, tx *models.Transaction) error
	
	// GetTransactionsByAccountID retrieves all transactions for a specific account
	GetTransactionsByAccountID(ctx context.Context, accountID string, page, pageSize int) ([]*models.Transaction, int64, error)
	
	// GetTransactionsByDateRange retrieves transactions within a date range
	GetTransactionsByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*models.Transaction, int64, error)
}
