package cte

import (
	"context"
	"time"
)

// EventState represents the state of a CTE event
type EventState string

const (
	// EventStateCreated indicates the event has been created but not yet started
	EventStateCreated EventState = "CREATED"
	// EventStateValidating indicates the event is being validated
	EventStateValidating EventState = "VALIDATING"
	// EventStateValidated indicates the event has been validated and is ready for execution
	EventStateValidated EventState = "VALIDATED"
	// EventStateExecuting indicates the event is currently being executed
	EventStateExecuting EventState = "EXECUTING"
	// EventStateCompleted indicates the event has completed successfully
	EventStateCompleted EventState = "COMPLETED"
	// EventStateFailed indicates the event has failed
	EventStateFailed EventState = "FAILED"
	// EventStateRollingBack indicates the event is being rolled back
	EventStateRollingBack EventState = "ROLLING_BACK"
	// EventStateRolledBack indicates the event has been rolled back
	EventStateRolledBack EventState = "ROLLED_BACK"
)

// TransactionState represents the state of a transaction within an event
type TransactionState string

const (
	// TransactionStatePending indicates the transaction is pending execution
	TransactionStatePending TransactionState = "PENDING"
	// TransactionStateExecuting indicates the transaction is currently being executed
	TransactionStateExecuting TransactionState = "EXECUTING"
	// TransactionStateCompleted indicates the transaction has completed successfully
	TransactionStateCompleted TransactionState = "COMPLETED"
	// TransactionStateFailed indicates the transaction has failed
	TransactionStateFailed TransactionState = "FAILED"
	// TransactionStateCompensating indicates the transaction is being compensated
	TransactionStateCompensating TransactionState = "COMPENSATING"
	// TransactionStateCompensated indicates the transaction has been compensated
	TransactionStateCompensated TransactionState = "COMPENSATED"
)

// Event represents a Chained Transaction Event
// This is the main entity that orchestrates the execution of multiple transactions
// as a single atomic unit of work.
type Event struct {
	// ID is the unique identifier for the event
	ID string `json:"id"`
	// Name is a human-readable name for the event
	Name string `json:"name"`
	// Description provides additional context about the event
	Description string `json:"description,omitempty"`
	// State is the current state of the event
	State EventState `json:"state"`
	// CreatedAt is the timestamp when the event was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the event was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// Timeout is the duration after which the event will time out
	Timeout time.Duration `json:"timeout,omitempty"`
	// Metadata contains additional context or parameters for the event
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Transaction represents a single transaction within an event
type Transaction struct {
	// ID is the unique identifier for the transaction
	ID string `json:"id"`
	// EventID is the ID of the event this transaction belongs to
	EventID string `json:"event_id"`
	// Name is a human-readable name for the transaction
	Name string `json:"name"`
	// Description provides additional context about the transaction
	Description string `json:"description,omitempty"`
	// Type indicates the type of transaction
	Type string `json:"type"`
	// State is the current state of the transaction
	State string `json:"state"`
	// Order defines the execution order of the transaction within the event
	Order int `json:"order"`
	// Dependencies is a list of transaction IDs that must complete before this transaction can start
	Dependencies []string `json:"dependencies,omitempty"`
	// Payload contains the data needed to execute the transaction
	Payload interface{} `json:"payload,omitempty"`
	// Result contains the result of the transaction execution
	Result interface{} `json:"result,omitempty"`
	// Error contains any error that occurred during transaction execution
	Error error `json:"error,omitempty"`
	// CreatedAt is the timestamp when the transaction was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the transaction was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// EventCoordinator orchestrates the execution of chained transaction events
type EventCoordinator interface {
	// CreateEvent creates a new CTE event
	CreateEvent(ctx context.Context, name, description string, timeout time.Duration, metadata map[string]interface{}) (*Event, error)
	// GetEvent retrieves an event by ID
	GetEvent(ctx context.Context, id string) (*Event, error)
	// AddTransaction adds a new transaction to an event
	AddTransaction(ctx context.Context, eventID string, tx *Transaction) error
	// StartEvent begins the execution of an event
	StartEvent(ctx context.Context, eventID string) error
	// GetEventTransactions retrieves all transactions for an event
	GetEventTransactions(ctx context.Context, eventID string) ([]*Transaction, error)
	// GetEventState retrieves the current state of an event
	GetEventState(ctx context.Context, eventID string) (EventState, error)
	// CompensateEvent triggers compensation for a failed event
	CompensateEvent(ctx context.Context, eventID string) error
}

// TransactionExecutor executes individual transactions within an event
type TransactionExecutor interface {
	// Execute executes a transaction
	Execute(ctx context.Context, tx *Transaction) error
	// Compensate compensates for a previously executed transaction
	Compensate(ctx context.Context, tx *Transaction) error
}

// EventStore persists the state of CTE events
type EventStore interface {
	// SaveEvent saves an event to the store
	SaveEvent(ctx context.Context, event *Event) error
	// GetEvent retrieves an event by ID
	GetEvent(ctx context.Context, id string) (*Event, error)
	// UpdateEvent updates an existing event
	UpdateEvent(ctx context.Context, event *Event) error
	// SaveTransaction saves a transaction to the store
	SaveTransaction(ctx context.Context, tx *Transaction) error
	// GetTransaction retrieves a transaction by ID
	GetTransaction(ctx context.Context, id string) (*Transaction, error)
	// UpdateTransaction updates an existing transaction
	UpdateTransaction(ctx context.Context, tx *Transaction) error
	// GetEventTransactions retrieves all transactions for an event
	GetEventTransactions(ctx context.Context, eventID string) ([]*Transaction, error)
}
