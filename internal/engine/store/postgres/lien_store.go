package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/engine/ctel"
	"gorm.io/gorm"
)

// LienModel represents the database model for CTE-liens
type LienModel struct {
	ID        string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EventID   string          `gorm:"type:uuid;not null;index"`
	AccountID string          `gorm:"type:uuid;not null;index"`
	Amount    float64         `gorm:"type:decimal(20,8);not null"`
	Currency  string          `gorm:"type:varchar(3);not null"`
	State     ctel.LienState  `gorm:"type:varchar(20);not null;default:'PENDING'"`
	ExpiresAt time.Time       `gorm:"not null"`
	Metadata  []byte          `gorm:"type:jsonb"`
	CreatedAt time.Time       `gorm:"not null;default:now()"`
	UpdatedAt time.Time       `gorm:"not null;default:now()"`
}

// TableName specifies the table name for the LienModel
func (LienModel) TableName() string {
	return "cte_liens"
}

// ToDomain converts the database model to a domain model
func (l *LienModel) ToDomain() (*ctel.Lien, error) {
	var metadata map[string]interface{}
	if len(l.Metadata) > 0 {
		if err := json.Unmarshal(l.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	return &ctel.Lien{
		ID:        l.ID,
		EventID:   l.EventID,
		AccountID: l.AccountID,
		Amount:    l.Amount,
		Currency:  l.Currency,
		State:     l.State,
		ExpiresAt: l.ExpiresAt,
		Metadata:  metadata,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
	}, nil
}

// FromDomain converts a domain model to a database model
func (l *LienModel) FromDomain(lien *ctel.Lien) error {
	l.ID = lien.ID
	l.EventID = lien.EventID
	l.AccountID = lien.AccountID
	l.Amount = lien.Amount
	l.Currency = lien.Currency
	l.State = lien.State
	l.ExpiresAt = lien.ExpiresAt
	l.CreatedAt = lien.CreatedAt
	l.UpdatedAt = lien.UpdatedAt

	if lien.Metadata != nil {
		metadata, err := json.Marshal(lien.Metadata)
		if err != nil {
			return err
		}
		l.Metadata = metadata
	}

	return nil
}

// LienStore implements the ctel.LienStore interface using PostgreSQL
type LienStore struct {
	db *gorm.DB
}

// NewLienStore creates a new PostgreSQL-based lien store
func NewLienStore(db *gorm.DB) *LienStore {
	return &LienStore{db: db}
}

// SaveLien saves a lien to the store
func (s *LienStore) SaveLien(ctx context.Context, lien *ctel.Lien) error {
	var model LienModel
	if err := model.FromDomain(lien); err != nil {
		return err
	}

	// Set updated timestamp
	model.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Create(&model).Error
}

// GetLien retrieves a lien by ID
func (s *LienStore) GetLien(ctx context.Context, id string) (*ctel.Lien, error) {
	var model LienModel
	if err := s.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain()
}

// GetLiensByEvent retrieves all liens for a specific CTE event
func (s *LienStore) GetLiensByEvent(ctx context.Context, eventID string) ([]*ctel.Lien, error) {
	var models []LienModel
	if err := s.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	liens := make([]*ctel.Lien, 0, len(models))
	for _, model := range models {
		lien, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		liens = append(liens, lien)
	}

	return liens, nil
}

// GetLiensByAccount retrieves all liens for a specific account
func (s *LienStore) GetLiensByAccount(ctx context.Context, accountID string) ([]*ctel.Lien, error) {
	var models []LienModel
	if err := s.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	liens := make([]*ctel.Lien, 0, len(models))
	for _, model := range models {
		lien, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		liens = append(liens, lien)
	}

	return liens, nil
}

// UpdateLien updates an existing lien
func (s *LienStore) UpdateLien(ctx context.Context, lien *ctel.Lien) error {
	var model LienModel
	if err := s.db.WithContext(ctx).First(&model, "id = ?", lien.ID).Error; err != nil {
		return err
	}

	if err := model.FromDomain(lien); err != nil {
		return err
	}

	// Set updated timestamp
	model.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Save(&model).Error
}

// Migrate creates the necessary database tables
func (s *LienStore) Migrate() error {
	return s.db.AutoMigrate(&LienModel{})
}
