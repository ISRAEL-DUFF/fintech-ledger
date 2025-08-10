package ctel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrLienNotFound is returned when a lien is not found
	ErrLienNotFound = errors.New("lien not found")
	// ErrInsufficientFunds is returned when there are insufficient funds to create a lien
	ErrInsufficientFunds = errors.New("insufficient funds")
	// ErrInvalidLienState is returned when a lien is in an invalid state for the requested operation
	ErrInvalidLienState = errors.New("invalid lien state for operation")
)

// LienManager implements the LienManager interface
type LienManager struct {
	store          LienStore
	accountService AccountService
}

// NewLienManager creates a new LienManager
func NewLienManager(store LienStore, accountService AccountService) *LienManager {
	return &LienManager{
		store:          store,
		accountService: accountService,
	}
}

// AccountService defines the interface for account operations needed by the LienManager
type AccountService interface {
	// GetAvailableBalance returns the available balance for an account
	GetAvailableBalance(ctx context.Context, accountID string) (float64, error)
}

// CreateLien creates a new lien for a CTE event
func (m *LienManager) CreateLien(
	ctx context.Context,
	eventID string,
	accountID string,
	amount float64,
	currency string,
	expiresAt time.Time,
	metadata map[string]interface{},
) (*Lien, error) {
	// Validate inputs
	if eventID == "" {
		return nil, errors.New("event ID is required")
	}
	if accountID == "" {
		return nil, errors.New("account ID is required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	if currency == "" {
		return nil, errors.New("currency is required")
	}
	if expiresAt.Before(time.Now()) {
		return nil, errors.New("expiration time must be in the future")
	}

	// Check if the account has sufficient funds
	available, err := m.accountService.GetAvailableBalance(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get available balance: %w", err)
	}

	// Get existing active liens for this account
	activeLiens, err := m.store.GetLiensByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active liens: %w", err)
	}

	// Calculate total reserved amount
	var reservedAmount float64
	for _, lien := range activeLiens {
		if lien.State == LienStateActive || lien.State == LienStatePending {
			reservedAmount += lien.Amount
		}
	}

	// Check if there are sufficient available funds
	if available-reservedAmount < amount {
		return nil, fmt.Errorf("%w: available=%.2f, requested=%.2f, reserved=%.2f",
			ErrInsufficientFunds, available, amount, reservedAmount)
	}

	// Create the lien
	lien := &Lien{
		ID:        uuid.New().String(),
		EventID:   eventID,
		AccountID: accountID,
		Amount:    amount,
		Currency:  currency,
		State:     LienStatePending,
		ExpiresAt: expiresAt,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save the lien
	if err := m.store.SaveLien(ctx, lien); err != nil {
		return nil, fmt.Errorf("failed to save lien: %w", err)
	}

	return lien, nil
}

// GetLien retrieves a lien by ID
func (m *LienManager) GetLien(ctx context.Context, id string) (*Lien, error) {
	lien, err := m.store.GetLien(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get lien: %w", err)
	}

	if lien == nil {
		return nil, ErrLienNotFound
	}

	return lien, nil
}

// GetLiensByEvent retrieves all liens for a specific CTE event
func (m *LienManager) GetLiensByEvent(ctx context.Context, eventID string) ([]*Lien, error) {
	liens, err := m.store.GetLiensByEvent(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get liens by event: %w", err)
	}

	return liens, nil
}

// GetLiensByAccount retrieves all liens for a specific account
func (m *LienManager) GetLiensByAccount(ctx context.Context, accountID string) ([]*Lien, error) {
	liens, err := m.store.GetLiensByAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get liens by account: %w", err)
	}

	return liens, nil
}

// ActivateLien activates a pending lien
func (m *LienManager) ActivateLien(ctx context.Context, id string) error {
	lien, err := m.GetLien(ctx, id)
	if err != nil {
		return err
	}

	// Only pending liens can be activated
	if lien.State != LienStatePending {
		return fmt.Errorf("%w: cannot activate lien in state %s",
			ErrInvalidLienState, lien.State)
	}

	// Check if the lien has expired
	if time.Now().After(lien.ExpiresAt) {
		lien.State = LienStateExpired
		lien.UpdatedAt = time.Now()

		if err := m.store.UpdateLien(ctx, lien); err != nil {
			return fmt.Errorf("failed to mark expired lien: %w", err)
		}

		return fmt.Errorf("lien has expired")
	}

	// Activate the lien
	lien.State = LienStateActive
	lien.UpdatedAt = time.Now()

	if err := m.store.UpdateLien(ctx, lien); err != nil {
		return fmt.Errorf("failed to activate lien: %w", err)
	}

	return nil
}

// ReleaseLien releases an active lien
func (m *LienManager) ReleaseLien(ctx context.Context, id string) error {
	lien, err := m.GetLien(ctx, id)
	if err != nil {
		return err
	}

	// Only active liens can be released
	if lien.State != LienStateActive {
		return fmt.Errorf("%w: cannot release lien in state %s",
			ErrInvalidLienState, lien.State)
	}

	// Mark the lien as released
	lien.State = LienStateReleased
	lien.UpdatedAt = time.Now()

	if err := m.store.UpdateLien(ctx, lien); err != nil {
		return fmt.Errorf("failed to release lien: %w", err)
	}

	return nil
}

// ExpireLien marks an expired lien as expired
func (m *LienManager) ExpireLien(ctx context.Context, id string) error {
	lien, err := m.GetLien(ctx, id)
	if err != nil {
		return err
	}

	// Only pending or active liens can be expired
	if lien.State != LienStatePending && lien.State != LienStateActive {
		return fmt.Errorf("%w: cannot expire lien in state %s",
			ErrInvalidLienState, lien.State)
	}

	// Mark the lien as expired
	lien.State = LienStateExpired
	lien.UpdatedAt = time.Now()

	if err := m.store.UpdateLien(ctx, lien); err != nil {
		return fmt.Errorf("failed to mark lien as expired: %w", err)
	}

	return nil
}

// GetAvailableBalance calculates the available balance for an account,
// taking into account active liens within the context of a CTE event
func (m *LienManager) GetAvailableBalance(
	ctx context.Context,
	eventID string,
	accountID string,
) (float64, error) {
	// Get the current available balance from the account service
	available, err := m.accountService.GetAvailableBalance(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("failed to get available balance: %w", err)
	}

	// Get all active liens for this account
	activeLiens, err := m.store.GetLiensByAccount(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("failed to get active liens: %w", err)
	}

	// Calculate total reserved amount from other events
	var reservedAmount float64
	for _, lien := range activeLiens {
		// Skip liens from the current event (they're already considered in the available balance)
		if lien.EventID == eventID {
			continue
		}

		if lien.State == LienStateActive || lien.State == LienStatePending {
			reservedAmount += lien.Amount
		}
	}

	// The available balance is the total available minus reserved amounts from other events
	return available - reservedAmount, nil
}
