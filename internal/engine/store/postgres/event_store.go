package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/engine/cte"
	"gorm.io/gorm"
)

// EventModel represents the database model for CTE events
type EventModel struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name        string         `gorm:"not null"`
	Description string         `gorm:"type:text"`
	State       cte.EventState `gorm:"type:varchar(20);not null;default:'CREATED'"`
	Timeout     *time.Duration `gorm:"type:interval"`
	Metadata    []byte         `gorm:"type:jsonb"`
	CreatedAt   time.Time      `gorm:"not null;default:now()"`
	UpdatedAt   time.Time      `gorm:"not null;default:now()"`
}

// TableName specifies the table name for the EventModel
func (EventModel) TableName() string {
	return "cte_events"
}

// ToDomain converts the database model to a domain model
func (e *EventModel) ToDomain() (*cte.Event, error) {
	var metadata map[string]interface{}
	if len(e.Metadata) > 0 {
		if err := json.Unmarshal(e.Metadata, &metadata); err != nil {
			return nil, err
		}
	}

	// Convert *time.Duration to time.Duration
	var timeout time.Duration
	if e.Timeout != nil {
		timeout = *e.Timeout
	}

	return &cte.Event{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		State:       e.State,
		Timeout:     timeout,
		Metadata:    metadata,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}, nil
}

// FromDomain converts a domain model to a database model
func (e *EventModel) FromDomain(event *cte.Event) error {
	e.ID = event.ID
	e.Name = event.Name
	e.Description = event.Description
	e.State = event.State
	
	// Convert time.Duration to *time.Duration
	if event.Timeout > 0 {
		timeout := event.Timeout
		e.Timeout = &timeout
	} else {
		e.Timeout = nil
	}
	
	e.CreatedAt = event.CreatedAt
	e.UpdatedAt = event.UpdatedAt

	if event.Metadata != nil {
		metadata, err := json.Marshal(event.Metadata)
		if err != nil {
			return err
		}
		e.Metadata = metadata
	}

	return nil
}

// TransactionModel represents the database model for CTE transactions
type TransactionModel struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EventID       string    `gorm:"type:uuid;not null;index"`
	Name          string    `gorm:"not null"`
	Description   string    `gorm:"type:text"`
	Type          string    `gorm:"not null"`
	State         string    `gorm:"type:varchar(20);not null;default:'PENDING'"`
	Order         int       `gorm:"not null"`
	Dependencies  []byte    `gorm:"type:jsonb"`
	Payload       []byte    `gorm:"type:jsonb"`
	Result        []byte    `gorm:"type:jsonb"`
	Error         string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"not null;default:now()"`
	UpdatedAt     time.Time `gorm:"not null;default:now()"`
}

// TableName specifies the table name for the TransactionModel
func (TransactionModel) TableName() string {
	return "cte_transactions"
}

// ToDomain converts the database model to a domain model
func (t *TransactionModel) ToDomain() (*cte.Transaction, error) {
	tx := &cte.Transaction{
		ID:          t.ID,
		EventID:     t.EventID,
		Name:        t.Name,
		Description: t.Description,
		Type:        t.Type,
		State:       t.State,
		Order:       t.Order,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}

	// Unmarshal dependencies
	if len(t.Dependencies) > 0 {
		var deps []string
		if err := json.Unmarshal(t.Dependencies, &deps); err != nil {
			return nil, err
		}
		tx.Dependencies = deps
	}

	// Unmarshal payload if present
	if len(t.Payload) > 0 {
		var payload interface{}
		if err := json.Unmarshal(t.Payload, &payload); err != nil {
			return nil, err
		}
		tx.Payload = payload
	}

	// Unmarshal result if present
	if len(t.Result) > 0 {
		var result interface{}
		if err := json.Unmarshal(t.Result, &result); err != nil {
			return nil, err
		}
		tx.Result = result
	}

	// Set error if present
	if t.Error != "" {
		tx.Error = errors.New(t.Error)
	}

	return tx, nil
}

// FromDomain converts a domain model to a database model
func (t *TransactionModel) FromDomain(tx *cte.Transaction) error {
	t.ID = tx.ID
	t.EventID = tx.EventID
	t.Name = tx.Name
	t.Description = tx.Description
	t.Type = tx.Type
	t.State = tx.State
	t.Order = tx.Order
	t.CreatedAt = tx.CreatedAt
	t.UpdatedAt = tx.UpdatedAt

	// Marshal dependencies
	if len(tx.Dependencies) > 0 {
		deps, err := json.Marshal(tx.Dependencies)
		if err != nil {
			return err
		}
		t.Dependencies = deps
	}

	// Marshal payload if present
	if tx.Payload != nil {
		payload, err := json.Marshal(tx.Payload)
		if err != nil {
			return err
		}
		t.Payload = payload
	}

	// Marshal result if present
	if tx.Result != nil {
		result, err := json.Marshal(tx.Result)
		if err != nil {
			return err
		}
		t.Result = result
	}

	// Set error message if present
	if tx.Error != nil {
		t.Error = tx.Error.Error()
	}

	return nil
}

// EventStore implements the cte.EventStore interface using PostgreSQL
type EventStore struct {
	db *gorm.DB
}

// NewEventStore creates a new PostgreSQL-based event store
func NewEventStore(db *gorm.DB) *EventStore {
	return &EventStore{db: db}
}

// SaveEvent saves an event to the store
func (s *EventStore) SaveEvent(ctx context.Context, event *cte.Event) error {
	var model EventModel
	if err := model.FromDomain(event); err != nil {
		return err
	}

	// Set updated timestamp
	model.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Save(&model).Error
}

// GetEvent retrieves an event by ID
func (s *EventStore) GetEvent(ctx context.Context, id string) (*cte.Event, error) {
	var model EventModel
	if err := s.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain()
}

// UpdateEvent updates an existing event
func (s *EventStore) UpdateEvent(ctx context.Context, event *cte.Event) error {
	var model EventModel
	if err := s.db.WithContext(ctx).First(&model, "id = ?", event.ID).Error; err != nil {
		return err
	}

	if err := model.FromDomain(event); err != nil {
		return err
	}

	// Set updated timestamp
	model.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Save(&model).Error
}

// SaveTransaction saves a transaction to the store
func (s *EventStore) SaveTransaction(ctx context.Context, tx *cte.Transaction) error {
	var model TransactionModel
	if err := model.FromDomain(tx); err != nil {
		return err
	}

	// Set updated timestamp
	model.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Save(&model).Error
}

// GetTransaction retrieves a transaction by ID
func (s *EventStore) GetTransaction(ctx context.Context, id string) (*cte.Transaction, error) {
	var model TransactionModel
	if err := s.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return model.ToDomain()
}

// UpdateTransaction updates an existing transaction
func (s *EventStore) UpdateTransaction(ctx context.Context, tx *cte.Transaction) error {
	var model TransactionModel
	if err := s.db.WithContext(ctx).First(&model, "id = ?", tx.ID).Error; err != nil {
		return err
	}

	if err := model.FromDomain(tx); err != nil {
		return err
	}

	// Set updated timestamp
	model.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Save(&model).Error
}

// GetEventTransactions retrieves all transactions for an event
func (s *EventStore) GetEventTransactions(ctx context.Context, eventID string) ([]*cte.Transaction, error) {
	var models []TransactionModel
	if err := s.db.WithContext(ctx).
		Where("event_id = ?", eventID).
		Order("\"order\" ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	transactions := make([]*cte.Transaction, 0, len(models))
	for _, model := range models {
		tx, err := model.ToDomain()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// Migrate creates the necessary database tables
func (s *EventStore) Migrate() error {
	return s.db.AutoMigrate(
		&EventModel{},
		&TransactionModel{},
	)
}
