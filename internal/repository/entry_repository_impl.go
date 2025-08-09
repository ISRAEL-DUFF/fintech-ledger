package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
	"gorm.io/gorm"
)

type entryRepository struct {
	db *gorm.DB
}

func NewEntryRepository(db *gorm.DB) EntryRepository {
	return &entryRepository{db: db}
}

func (r *entryRepository) CreateEntry(ctx context.Context, entry *models.Entry) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Generate a new UUID for the entry if not set
		if entry.ID == "" {
			entry.ID = uuid.New().String()
		}

		// Create the entry
		if err := tx.Create(entry).Error; err != nil {
			return err
		}
		
		// Create all entry lines
		for i := range entry.Lines {
			entry.Lines[i].ID = uuid.New().String()
			entry.Lines[i].EntryID = entry.ID
			entry.Lines[i].CreatedAt = time.Now()

			if err := tx.Create(&entry.Lines[i]).Error; err != nil {
				return err
			}
		}
		
		return nil
	})
}

func (r *entryRepository) GetEntryByID(ctx context.Context, id string) (*models.Entry, error) {
	var entry models.Entry
	err := r.db.WithContext(ctx).
		Preload("Lines").
		First(&entry, "id = ?", id).
		Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

func (r *entryRepository) GetEntriesByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*models.Entry, int64, error) {
	var entries []*models.Entry
	var total int64

	// First, get the total count
	err := r.db.WithContext(ctx).
		Model(&models.Entry{}).
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Count(&total).
		Error

	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = r.db.WithContext(ctx).
		Preload("Lines").
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&entries).
		Error

	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}
