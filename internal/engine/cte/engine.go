package cte

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrEventNotFound is returned when an event is not found
	ErrEventNotFound = errors.New("event not found")
	// ErrTransactionNotFound is returned when a transaction is not found
	ErrTransactionNotFound = errors.New("transaction not found")
	// ErrInvalidEventState is returned when an event is in an invalid state for the requested operation
	ErrInvalidEventState = errors.New("invalid event state for operation")
	// ErrTransactionDependencyNotMet is returned when a transaction's dependencies are not met
	ErrTransactionDependencyNotMet = errors.New("transaction dependencies not met")
)

// Engine implements the EventCoordinator interface
type Engine struct {
	eventStore  EventStore
	txExecutors map[string]TransactionExecutor
	maxRetries  int
	retryDelay  time.Duration
	mu          sync.RWMutex
}

// NewEngine creates a new CTE engine
func NewEngine(eventStore EventStore) *Engine {
	return &Engine{
		eventStore:  eventStore,
		txExecutors: make(map[string]TransactionExecutor),
		maxRetries:  3,
		retryDelay:  100 * time.Millisecond,
	}
}

// RegisterExecutor registers a transaction executor for a specific transaction type
func (e *Engine) RegisterExecutor(txType string, executor TransactionExecutor) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.txExecutors[txType] = executor
}

// CreateEvent creates a new CTE event
func (e *Engine) CreateEvent(ctx context.Context, name, description string, timeout time.Duration, metadata map[string]interface{}) (*Event, error) {
	event := &Event{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		State:       EventStateCreated,
		Timeout:     timeout,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := e.eventStore.SaveEvent(ctx, event); err != nil {
		return nil, fmt.Errorf("failed to save event: %w", err)
	}

	return event, nil
}

// GetEvent retrieves an event by ID
func (e *Engine) GetEvent(ctx context.Context, id string) (*Event, error) {
	event, err := e.eventStore.GetEvent(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	if event == nil {
		return nil, ErrEventNotFound
	}

	return event, nil
}

// AddTransaction adds a new transaction to an event
func (e *Engine) AddTransaction(ctx context.Context, eventID string, tx *Transaction) error {
	// Get the event to verify it exists and is in a valid state
	event, err := e.GetEvent(ctx, eventID)
	if err != nil {
		return err
	}

	// Only allow adding transactions to events that haven't started yet
	if event.State != EventStateCreated && event.State != EventStateValidating {
		return fmt.Errorf("%w: cannot add transaction to event in state %s",
			ErrInvalidEventState, event.State)
	}

	// Generate a new ID if not provided
	if tx.ID == "" {
		tx.ID = uuid.New().String()
	}

	// Set timestamps
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()

	// Save the transaction
	if err := e.eventStore.SaveTransaction(ctx, tx); err != nil {
		return fmt.Errorf("failed to save transaction: %w", err)
	}

	// Update event state to validating if this is the first transaction
	if event.State == EventStateCreated {
		event.State = EventStateValidating
		if err := e.eventStore.UpdateEvent(ctx, event); err != nil {
			return fmt.Errorf("failed to update event state: %w", err)
		}
	}

	return nil
}

// StartEvent begins the execution of an event
func (e *Engine) StartEvent(ctx context.Context, eventID string) error {
	event, err := e.GetEvent(ctx, eventID)
	if err != nil {
		return err
	}

	// Only allow starting events that are in the VALIDATED state
	if event.State != EventStateValidated {
		return fmt.Errorf("%w: cannot start event in state %s",
			ErrInvalidEventState, event.State)
	}

	// Update event state to EXECUTING
	event.State = EventStateExecuting
	event.UpdatedAt = time.Now()
	if err := e.eventStore.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to update event state: %w", err)
	}

	// Start executing transactions in a separate goroutine
	go e.executeEvent(context.Background(), eventID)

	return nil
}

// executeEvent executes all transactions in an event
func (e *Engine) executeEvent(ctx context.Context, eventID string) error {
	// Get all transactions for the event
	transactions, err := e.eventStore.GetEventTransactions(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event transactions: %w", err)
	}

	// Execute transactions in order, respecting dependencies
	for _, tx := range transactions {
		// Check if all dependencies are met
		if err := e.checkDependencies(ctx, tx); err != nil {
			if errors.Is(err, ErrTransactionDependencyNotMet) {
				// Skip this transaction for now, it will be retried later
				continue
			}
			// For other errors, mark the event as failed
			return e.failEvent(ctx, eventID, fmt.Errorf("failed to check dependencies: %w", err))
		}

		// Execute the transaction with retries
		if err := e.executeTransactionWithRetry(ctx, tx); err != nil {
			// If execution fails, trigger compensation
			if compErr := e.compensateEvent(ctx, eventID); compErr != nil {
				return fmt.Errorf("failed to compensate event after transaction failure: %v (original error: %w)",
					compErr, err)
			}
			return err
		}
	}

	// If we get here, all transactions completed successfully
	event, err := e.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event for completion: %w", err)
	}

	event.State = EventStateCompleted
	event.UpdatedAt = time.Now()
	if err := e.eventStore.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to mark event as completed: %w", err)
	}

	return nil
}

// executeTransactionWithRetry executes a transaction with retries
func (e *Engine) executeTransactionWithRetry(ctx context.Context, tx *Transaction) error {
	var lastErr error

	for attempt := 0; attempt < e.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(e.retryDelay)
		}

		tx.State = "EXECUTING"
		tx.UpdatedAt = time.Now()
		if err := e.eventStore.UpdateTransaction(ctx, tx); err != nil {
			lastErr = fmt.Errorf("failed to update transaction state: %w", err)
			continue
		}

		// Get the executor for this transaction type
		e.mu.RLock()
		executor, ok := e.txExecutors[tx.Type]
		e.mu.RUnlock()

		if !ok {
			lastErr = fmt.Errorf("no executor registered for transaction type: %s", tx.Type)
			continue
		}

		// Execute the transaction
		if err := executor.Execute(ctx, tx); err != nil {
			lastErr = err
			tx.State = "FAILED"
			tx.Error = err
			tx.UpdatedAt = time.Now()
			if updateErr := e.eventStore.UpdateTransaction(ctx, tx); updateErr != nil {
				// Log the error but continue with the original error
				lastErr = fmt.Errorf("failed to update failed transaction: %v (original error: %w)",
					updateErr, lastErr)
			}
			continue
		}

		// If we get here, the transaction was successful
		tx.State = "COMPLETED"
		tx.UpdatedAt = time.Now()
		if err := e.eventStore.UpdateTransaction(ctx, tx); err != nil {
			return fmt.Errorf("failed to update completed transaction: %w", err)
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// checkDependencies verifies that all of a transaction's dependencies are met
func (e *Engine) checkDependencies(ctx context.Context, tx *Transaction) error {
	if len(tx.Dependencies) == 0 {
		return nil // No dependencies to check
	}

	for _, depID := range tx.Dependencies {
		depTx, err := e.eventStore.GetTransaction(ctx, depID)
		if err != nil {
			return fmt.Errorf("failed to get dependency transaction %s: %w", depID, err)
		}

		if depTx == nil {
			return fmt.Errorf("dependency transaction %s not found", depID)
		}

		if depTx.State != "COMPLETED" {
			return fmt.Errorf("%w: dependency %s is in state %s",
				ErrTransactionDependencyNotMet, depID, depTx.State)
		}
	}

	return nil
}

// failEvent marks an event as failed
func (e *Engine) failEvent(ctx context.Context, eventID string, err error) error {
	event, err2 := e.GetEvent(ctx, eventID)
	if err2 != nil {
		return fmt.Errorf("failed to get event for failure: %v (original error: %w)", err2, err)
	}

	event.State = EventStateFailed
	event.UpdatedAt = time.Now()
	if updateErr := e.eventStore.UpdateEvent(ctx, event); updateErr != nil {
		return fmt.Errorf("failed to mark event as failed: %v (original error: %w)", updateErr, err)
	}

	return err
}

// compensateEvent compensates for all completed transactions in an event
func (e *Engine) compensateEvent(ctx context.Context, eventID string) error {
	// Mark the event as rolling back
	event, err := e.GetEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event for compensation: %w", err)
	}

	event.State = EventStateRollingBack
	event.UpdatedAt = time.Now()
	if err := e.eventStore.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to update event state to rolling back: %w", err)
	}

	// Get all transactions for the event
	transactions, err := e.eventStore.GetEventTransactions(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get event transactions for compensation: %w", err)
	}

	// Execute compensation in reverse order
	for i := len(transactions) - 1; i >= 0; i-- {
		tx := transactions[i]

		// Only compensate completed transactions
		if tx.State != "COMPLETED" {
			continue
		}

		// Get the executor for this transaction type
		e.mu.RLock()
		executor, ok := e.txExecutors[tx.Type]
		e.mu.RUnlock()

		if !ok {
			// Log the error but continue with other transactions
			continue
		}

		// Execute the compensation
		if err := executor.Compensate(ctx, tx); err != nil {
			// Log the error but continue with other transactions
			continue
		}

		// Mark the transaction as compensated
		tx.State = "COMPENSATED"
		tx.UpdatedAt = time.Now()
		if err := e.eventStore.UpdateTransaction(ctx, tx); err != nil {
			// Log the error but continue with other transactions
			continue
		}
	}

	// Mark the event as rolled back
	event.State = EventStateRolledBack
	event.UpdatedAt = time.Now()
	if err := e.eventStore.UpdateEvent(ctx, event); err != nil {
		return fmt.Errorf("failed to mark event as rolled back: %w", err)
	}

	return nil
}

// GetEventTransactions retrieves all transactions for an event
func (e *Engine) GetEventTransactions(ctx context.Context, eventID string) ([]*Transaction, error) {
	transactions, err := e.eventStore.GetEventTransactions(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event transactions: %w", err)
	}

	return transactions, nil
}

// GetEventState retrieves the current state of an event
func (e *Engine) GetEventState(ctx context.Context, eventID string) (EventState, error) {
	event, err := e.GetEvent(ctx, eventID)
	if err != nil {
		return "", err
	}

	return event.State, nil
}

// CompensateEvent triggers compensation for a failed event
func (e *Engine) CompensateEvent(ctx context.Context, eventID string) error {
	return e.compensateEvent(ctx, eventID)
}
