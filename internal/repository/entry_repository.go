package repository

import (
	"context"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
)

// EntryRepository defines the interface for entry data operations
type EntryRepository interface {
	CreateEntry(ctx context.Context, entry *models.Entry) error
	GetEntryByID(ctx context.Context, id string) (*models.Entry, error)
	GetEntriesByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*models.Entry, int64, error)
}
