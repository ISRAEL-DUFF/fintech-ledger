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

// WalletDepositPayload defines the structure for wallet deposit transaction payload
type WalletDepositPayload struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Reference string  `json:"reference,omitempty"`
	Source    string  `json:"source,omitempty"`
}

// WalletDepositResult defines the structure for wallet deposit transaction result
type WalletDepositResult struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	ProcessedAt   time.Time `json:"processed_at"`
}

// WalletDepositExecutor handles wallet deposit transactions
type WalletDepositExecutor struct {
	db             *gorm.DB
	accountRepo    repository.AccountRepository
	transactionSvc service.TransactionService
}

// NewWalletDepositExecutor creates a new wallet deposit executor
func NewWalletDepositExecutor(
	db *gorm.DB,
	accountRepo repository.AccountRepository,
	transactionSvc service.TransactionService,
) *WalletDepositExecutor {
	return &WalletDepositExecutor{
		db:             db,
		accountRepo:    accountRepo,
		transactionSvc: transactionSvc,
	}
}

// Execute processes a wallet deposit transaction
func (e *WalletDepositExecutor) Execute(ctx context.Context, tx *cte.Transaction) error {
	// Parse the payload
	var payload WalletDepositPayload
	payloadBytes, err := json.Marshal(tx.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Validate the payload
	if err := validateWalletDepositPayload(&payload); err != nil {
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

	// Get the account
	account, err := e.accountRepo.GetAccountByID(ctx, payload.AccountID)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to get account: %w", err)
	}

	// Check if account supports the specified currency
	if !accountSupportsCurrency(account, payload.Currency) {
		dbTx.Rollback()
		return fmt.Errorf("account does not support currency %s", payload.Currency)
	}

	// Process the deposit using the transaction service
	depositReq := service.DepositRequest{
		AccountID: payload.AccountID,
		Amount:    payload.Amount,
		Currency:  payload.Currency,
		Reference: payload.Reference,
	}

	// Add source to reference if provided
	if payload.Source != "" {
		if depositReq.Reference != "" {
			depositReq.Reference = fmt.Sprintf("%s (Source: %s)", depositReq.Reference, payload.Source)
		} else {
			depositReq.Reference = fmt.Sprintf("Deposit from %s", payload.Source)
		}
	}

	_, err = e.transactionSvc.ProcessDeposit(ctx, depositReq)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to process deposit: %w", err)
	}

	// Commit the transaction
	if commitErr := dbTx.Commit().Error; commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

// Compensate handles the rollback of a wallet deposit
func (e *WalletDepositExecutor) Compensate(ctx context.Context, tx *cte.Transaction) error {
	// Parse the payload
	var payload WalletDepositPayload
	payloadBytes, err := json.Marshal(tx.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Get the transaction ID from the result if available
	var result WalletDepositResult
	if tx.Result != nil {
		resultBytes, err := json.Marshal(tx.Result)
		if err == nil {
			json.Unmarshal(resultBytes, &result)
		}
	}

	// If we have a transaction ID, try to reverse it
	if result.TransactionID != "" {
		if err := e.transactionSvc.ReverseDeposit(ctx, result.TransactionID); err != nil {
			return fmt.Errorf("failed to reverse deposit: %w", err)
		}
		return nil
	}

	// If we don't have a transaction ID, manually reverse the deposit
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

	// Create a withdrawal to reverse the deposit
	withdrawalReq := service.WithdrawalRequest{
		AccountID: payload.AccountID,
		Amount:    payload.Amount,
		Currency:  payload.Currency,
		Reference: fmt.Sprintf("REV-%s", payload.Reference),
	}

	// Add source to reference if provided
	if payload.Source != "" {
		if withdrawalReq.Reference != "" {
			withdrawalReq.Reference = fmt.Sprintf("%s (Original source: %s)", withdrawalReq.Reference, payload.Source)
		} else {
			withdrawalReq.Reference = fmt.Sprintf("Reversal of deposit from %s", payload.Source)
		}
	}

	// Process the withdrawal to reverse the deposit
	_, err = e.transactionSvc.ProcessWithdrawal(ctx, withdrawalReq)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to process withdrawal to reverse deposit: %w", err)
	}

	// Commit the transaction
	if commitErr := dbTx.Commit().Error; commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

// validateWalletDepositPayload validates the wallet deposit payload
func validateWalletDepositPayload(payload *WalletDepositPayload) error {
	if payload.AccountID == "" {
		return fmt.Errorf("account ID is required")
	}

	if payload.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}

	if payload.Currency == "" {
		return fmt.Errorf("currency is required")
	}

	return nil
}


