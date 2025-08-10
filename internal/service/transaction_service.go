package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/repository"
	"github.com/google/uuid"
)

// transactionServiceImpl is the implementation of TransactionService
type transactionServiceImpl struct {
	repo         repository.EntryRepository
	accountRepo  repository.AccountRepository
}

// TransactionResponse represents the response for transaction operations
type TransactionResponse struct {
	ID              string    `json:"id"`
	Description     string    `json:"description"`
	TransactionType string    `json:"transaction_type"`
	ReferenceID     string    `json:"reference_id,omitempty"`
	Status          string    `json:"status"`
	Date            time.Time `json:"date"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Lines           []LineResponse `json:"lines,omitempty"`
}

type LineResponse struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	Debit     float64   `json:"debit"`
	Credit    float64   `json:"credit"`
	CreatedAt time.Time `json:"created_at"`
}

// NewTransactionService creates a new TransactionService
func NewTransactionService(entryRepo repository.EntryRepository, accountRepo repository.AccountRepository) TransactionService {
	return &transactionServiceImpl{
		repo:         entryRepo,
		accountRepo:  accountRepo,
	}
}

// ValidateEntry ensures the entry follows double-entry accounting rules
func (s *transactionServiceImpl) ValidateEntry(ctx context.Context, entry *models.Entry) error {
	if len(entry.Lines) < 2 {
		return errors.New("entry must have at least two lines")
	}

	var totalDebit, totalCredit float64
	accountIDs := make(map[string]bool)

	for _, line := range entry.Lines {
		// Ensure account exists
		if _, exists := accountIDs[line.AccountID]; !exists {
			account, err := s.accountRepo.GetAccountByID(ctx, line.AccountID)
			if err != nil {
				return fmt.Errorf("error validating account %s: %w", line.AccountID, err)
			}
			if account == nil {
				return fmt.Errorf("account %s not found", line.AccountID)
			}
			accountIDs[line.AccountID] = true
		}

		// Validate line amounts
		if line.Debit < 0 || line.Credit < 0 {
			return errors.New("debit and credit amounts must be non-negative")
		}
		if line.Debit > 0 && line.Credit > 0 {
			return errors.New("a line cannot have both debit and credit amounts")
		}

		totalDebit += line.Debit
		totalCredit += line.Credit
	}

	if totalDebit != totalCredit {
		return fmt.Errorf("total debits (%f) do not equal total credits (%f)", totalDebit, totalCredit)
	}

	return nil
}

// CreateEntry creates a new transaction entry with validation
func (s *transactionServiceImpl) CreateEntry(ctx context.Context, entry *models.Entry) error {
	// Validate the entry
	if err := s.ValidateEntry(ctx, entry); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set timestamps
	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now

	// Set timestamps for entry lines
	for i := range entry.Lines {
		entry.Lines[i].CreatedAt = now
	}

	// Generate entry ID if not set
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// Generate line IDs and set entry ID reference
	for i := range entry.Lines {
		entry.Lines[i].ID = uuid.New().String()
		entry.Lines[i].EntryID = entry.ID
	}

	return s.repo.CreateEntry(ctx, entry)
}

// GetEntryByID retrieves a transaction entry by its ID
func (s *transactionServiceImpl) GetEntryByID(ctx context.Context, id string) (*models.Entry, error) {
	return s.repo.GetEntryByID(ctx, id)
}

// GetEntriesByDateRange retrieves transaction entries within a date range with pagination
func (s *transactionServiceImpl) GetEntriesByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*models.Entry, int64, error) {
	return s.repo.GetEntriesByDateRange(ctx, startDate, endDate, page, pageSize)
}

// ProcessTransfer processes a transfer between two accounts
func (s *transactionServiceImpl) ProcessTransfer(ctx context.Context, req TransferRequest) (*models.Transaction, error) {
	// TODO: Implement transfer logic with proper validation and transaction handling
	tx := &models.Transaction{
		ID:          uuid.New().String(),
		Type:        "transfer",
		Status:      "completed",
		Description: fmt.Sprintf("Transfer of %f %s to account %s", req.Amount, req.Currency, req.DestinationAccountID),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return tx, nil
}

// ProcessDeposit processes a deposit to an account
func (s *transactionServiceImpl) ProcessDeposit(ctx context.Context, req DepositRequest) (*models.Transaction, error) {
	// TODO: Implement deposit logic with proper validation and transaction handling
	tx := &models.Transaction{
		ID:          uuid.New().String(),
		Type:        "deposit",
		Status:      "completed",
		Description: fmt.Sprintf("Deposit of %f %s", req.Amount, req.Currency),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return tx, nil
}

// ProcessWithdrawal processes a withdrawal from an account
func (s *transactionServiceImpl) ProcessWithdrawal(ctx context.Context, req WithdrawalRequest) (*models.Transaction, error) {
	// TODO: Implement withdrawal logic with proper validation and transaction handling
	tx := &models.Transaction{
		ID:          uuid.New().String(),
		Type:        "withdrawal",
		Status:      "completed",
		Description: fmt.Sprintf("Withdrawal of %f %s", req.Amount, req.Currency),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return tx, nil
}

// ProcessExchange processes a currency exchange between two accounts
func (s *transactionServiceImpl) ProcessExchange(ctx context.Context, req ExchangeRequest) (*models.Transaction, error) {
	// TODO: Implement currency exchange logic with proper validation and transaction handling
	tx := &models.Transaction{
		ID:          uuid.New().String(),
		Type:        "exchange",
		Status:      "completed",
		Description: fmt.Sprintf("Exchange %f %s to %f %s", req.SourceAmount, req.SourceCurrency, req.DestinationAmount, req.DestinationCurrency),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return tx, nil
}

// ProcessFee processes a fee transaction
func (s *transactionServiceImpl) ProcessFee(ctx context.Context, req FeeRequest) (*models.Transaction, error) {
	// TODO: Implement fee processing logic
	// This is a placeholder implementation
	tx := &models.Transaction{
		ID:          uuid.New().String(),
		Type:        "fee",
		Status:      "completed",
		Description: fmt.Sprintf("Fee of %f %s", req.Amount, req.Currency),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	return tx, nil
}

// ReverseTransfer reverses a transfer transaction
func (s *transactionServiceImpl) ReverseTransfer(ctx context.Context, transactionID string) error {
	// TODO: Implement transfer reversal logic with proper validation and transaction handling
	return nil
}

// ReverseDeposit reverses a deposit transaction
func (s *transactionServiceImpl) ReverseDeposit(ctx context.Context, transactionID string) error {
	// TODO: Implement deposit reversal logic with proper validation and transaction handling
	return nil
}

// ReverseWithdrawal reverses a withdrawal transaction
func (s *transactionServiceImpl) ReverseWithdrawal(ctx context.Context, transactionID string) error {
	// TODO: Implement withdrawal reversal logic with proper validation and transaction handling
	return nil
}

// ReverseExchange reverses a currency exchange transaction
func (s *transactionServiceImpl) ReverseExchange(ctx context.Context, transactionID string) error {
	// TODO: Implement exchange reversal logic with proper validation and transaction handling
	return nil
}

// ReverseFee reverses a fee transaction
func (s *transactionServiceImpl) ReverseFee(ctx context.Context, transactionID string) error {
	// TODO: Implement fee reversal logic with proper validation and transaction handling
	return nil
}
