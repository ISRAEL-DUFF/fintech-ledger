package service

import (
	"context"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
)

// TransactionService defines the interface for transaction operations
type TransactionService interface {
	CreateEntry(ctx context.Context, entry *models.Entry) error
	GetEntryByID(ctx context.Context, id string) (*models.Entry, error)
	GetEntriesByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*models.Entry, int64, error)
	ValidateEntry(ctx context.Context, entry *models.Entry) error
}
