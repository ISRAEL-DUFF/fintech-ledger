package executors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/engine/cte"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/repository"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/service"
	"gorm.io/gorm"
)

// WalletTransferPayload defines the structure for wallet transfer transaction payload
type WalletTransferPayload struct {
	SourceAccountID      string  `json:"source_account_id"`
	DestinationAccountID string  `json:"destination_account_id"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	Reference            string  `json:"reference,omitempty"`
}

// WalletTransferResult defines the structure for wallet transfer transaction result
type WalletTransferResult struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	ProcessedAt   time.Time `json:"processed_at"`
}

// WalletTransferExecutor handles wallet transfer transactions
type WalletTransferExecutor struct {
	db             *gorm.DB
	accountRepo    repository.AccountRepository
	transactionSvc service.TransactionService
}

// NewWalletTransferExecutor creates a new wallet transfer executor
func NewWalletTransferExecutor(
	db *gorm.DB,
	accountRepo repository.AccountRepository,
	transactionSvc service.TransactionService,
) *WalletTransferExecutor {
	return &WalletTransferExecutor{
		db:             db,
		accountRepo:    accountRepo,
		transactionSvc: transactionSvc,
	}
}

// Execute processes a wallet transfer transaction
func (e *WalletTransferExecutor) Execute(ctx context.Context, tx *cte.Transaction) error {
	// Parse the payload
	var payload WalletTransferPayload
	payloadBytes, err := json.Marshal(tx.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Validate the payload
	if err := validateWalletTransferPayload(&payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	// Begin a database transaction
	dbTx := e.db.Begin()
	if dbTx.Error != nil {
		return fmt.Errorf("failed to begin database transaction: %w", dbTx.Error)
	}

	// Defer a rollback in case anything fails
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()

	// Get the source and destination accounts
	sourceAccount, err := e.accountRepo.GetAccountByID(ctx, payload.SourceAccountID)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to get source account: %w", err)
	}

	destAccount, err := e.accountRepo.GetAccountByID(ctx, payload.DestinationAccountID)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to get destination account: %w", err)
	}

	// Check if accounts support the specified currency
	if !accountSupportsCurrency(sourceAccount, payload.Currency) {
		dbTx.Rollback()
		return fmt.Errorf("source account does not support currency %s", payload.Currency)
	}

	if !accountSupportsCurrency(destAccount, payload.Currency) {
		dbTx.Rollback()
		return fmt.Errorf("destination account does not support currency %s", payload.Currency)
	}

	// Process the transfer using the transaction service
	transferReq := service.TransferRequest{
		SourceAccountID:      payload.SourceAccountID,
		DestinationAccountID: payload.DestinationAccountID,
		Amount:               payload.Amount,
		Currency:             payload.Currency,
		Reference:            payload.Reference,
	}

	_, err = e.transactionSvc.ProcessTransfer(ctx, transferReq)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to process transfer: %w", err)
	}

	// Commit the transaction
	if commitErr := dbTx.Commit().Error; commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

// Compensate handles the rollback of a wallet transfer
func (e *WalletTransferExecutor) Compensate(ctx context.Context, tx *cte.Transaction) error {
	// Parse the payload
	var payload WalletTransferPayload
	payloadBytes, err := json.Marshal(tx.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Get the transaction ID from the result if available
	var result WalletTransferResult
	if tx.Result != nil {
		resultBytes, err := json.Marshal(tx.Result)
		if err == nil {
			json.Unmarshal(resultBytes, &result)
		}
	}

	// If we have a transaction ID, try to reverse it
	if result.TransactionID != "" {
		if err := e.transactionSvc.ReverseTransfer(ctx, result.TransactionID); err != nil {
			return fmt.Errorf("failed to reverse transfer: %w", err)
		}
		return nil
	}

	// If we don't have a transaction ID, manually reverse the transfer
	dbTx := e.db.Begin()
	if dbTx.Error != nil {
		return fmt.Errorf("failed to begin database transaction: %w", dbTx.Error)
	}

	// Defer a rollback in case anything fails
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
		}
	}()

	// Create a reverse transfer request
	reverseReq := service.TransferRequest{
		SourceAccountID:      payload.DestinationAccountID, // Reverse source and target
		DestinationAccountID: payload.SourceAccountID,
		Amount:               payload.Amount,
		Currency:             payload.Currency,
		Reference:            fmt.Sprintf("REV-%s", payload.Reference),
	}

	// Process the reverse transfer
	_, err = e.transactionSvc.ProcessTransfer(ctx, reverseReq)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to process reverse transfer: %w", err)
	}

	// Commit the transaction
	if commitErr := dbTx.Commit().Error; commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

// validateWalletTransferPayload validates the wallet transfer payload
func validateWalletTransferPayload(payload *WalletTransferPayload) error {
	if payload.SourceAccountID == "" {
		return fmt.Errorf("source account ID is required")
	}

	if payload.DestinationAccountID == "" {
		return fmt.Errorf("destination account ID is required")
	}

	if payload.SourceAccountID == payload.DestinationAccountID {
		return fmt.Errorf("source and destination accounts cannot be the same")
	}

	if payload.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	if payload.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	return nil
}
