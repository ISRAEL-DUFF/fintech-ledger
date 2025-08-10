package executors

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/engine/cte"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/engine/ctel"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/repository"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/service"
	"gorm.io/gorm"
)

// WalletWithdrawalPayload defines the structure for wallet withdrawal transaction payload
type WalletWithdrawalPayload struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Reference string  `json:"reference,omitempty"`
	Target    string  `json:"target,omitempty"`
}

// WalletWithdrawalResult defines the structure for wallet withdrawal transaction result
type WalletWithdrawalResult struct {
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	ProcessedAt   time.Time `json:"processed_at"`
}

// WalletWithdrawalExecutor handles wallet withdrawal transactions
type WalletWithdrawalExecutor struct {
	db             *gorm.DB
	accountRepo    repository.AccountRepository
	transactionSvc service.TransactionService
	lienManager    ctel.LienManager
}

// NewWalletWithdrawalExecutor creates a new wallet withdrawal executor
func NewWalletWithdrawalExecutor(
	db *gorm.DB,
	accountRepo repository.AccountRepository,
	transactionSvc service.TransactionService,
	lienManager ctel.LienManager,
) *WalletWithdrawalExecutor {
	return &WalletWithdrawalExecutor{
		db:             db,
		accountRepo:    accountRepo,
		transactionSvc: transactionSvc,
		lienManager:    lienManager,
	}
}

// Execute processes a wallet withdrawal transaction
func (e *WalletWithdrawalExecutor) Execute(ctx context.Context, tx *cte.Transaction) error {
	// Parse the payload
	var payload WalletWithdrawalPayload
	payloadBytes, err := json.Marshal(tx.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Validate the payload
	if err := validateWalletWithdrawalPayload(&payload); err != nil {
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

	// Create a lien to reserve the funds
	lien, err := e.lienManager.CreateLien(
		ctx,
		tx.EventID,
		payload.AccountID,
		payload.Amount,
		payload.Currency,
		time.Now().Add(30*time.Minute), // 30-minute lien expiration
		nil,
	)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to create lien: %w", err)
	}

	// Create withdrawal request
	withdrawalReq := service.WithdrawalRequest{
		AccountID: payload.AccountID,
		Amount:    payload.Amount,
		Currency:  payload.Currency,
		Reference: payload.Reference,
	}

	// Process the withdrawal using the transaction service
	transaction, err := e.transactionSvc.ProcessWithdrawal(ctx, withdrawalReq)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to process withdrawal: %w", err)
	}

	// Release the lien since the withdrawal was successful
	if err := e.lienManager.ReleaseLien(ctx, lien.ID); err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to release lien: %w", err)
	}

	// Create the result
	result := WalletWithdrawalResult{
		TransactionID: transaction.ID,
		Status:        "COMPLETED",
		ProcessedAt:   time.Now(),
	}

	// Convert result to map for storage
	resultBytes, err := json.Marshal(result)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(resultBytes, &resultMap); err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	tx.Result = resultMap
	tx.UpdatedAt = time.Now()

	// Commit the transaction
	if commitErr := dbTx.Commit().Error; commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

// Compensate handles the rollback of a wallet withdrawal
func (e *WalletWithdrawalExecutor) Compensate(ctx context.Context, tx *cte.Transaction) error {
	// Parse the payload
	var payload WalletWithdrawalPayload
	payloadBytes, err := json.Marshal(tx.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Parse the result to get the transaction ID
	var result WalletWithdrawalResult
	if tx.Result != nil {
		resultBytes, err := json.Marshal(tx.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		if err := json.Unmarshal(resultBytes, &result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
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

	// If the transaction was completed, we need to reverse it
	if result.Status == "COMPLETED" {
		// Reverse the withdrawal
		if err := e.transactionSvc.ReverseWithdrawal(ctx, result.TransactionID); err != nil {
			dbTx.Rollback()
			return fmt.Errorf("failed to reverse withdrawal: %w", err)
		}

		// Update the result
		result.Status = "REVERSED"
		result.ProcessedAt = time.Now()

		resultBytes, err := json.Marshal(result)
		if err != nil {
			dbTx.Rollback()
			return fmt.Errorf("failed to marshal result: %w", err)
		}

		var resultMap map[string]interface{}
		if err := json.Unmarshal(resultBytes, &resultMap); err != nil {
			dbTx.Rollback()
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}

		tx.Result = resultMap
		tx.UpdatedAt = time.Now()
	}

	// Release any active liens for this event
	liens, err := e.lienManager.GetLiensByEvent(ctx, tx.EventID)
	if err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to get liens for event: %w", err)
	}

	for _, lien := range liens {
		if lien.State == ctel.LienStateActive || lien.State == ctel.LienStatePending {
			if err := e.lienManager.ReleaseLien(ctx, lien.ID); err != nil {
				dbTx.Rollback()
				return fmt.Errorf("failed to release lien %s: %w", lien.ID, err)
			}
		}
	}

	// Commit the transaction
	if commitErr := dbTx.Commit().Error; commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

// validateWalletWithdrawalPayload validates the wallet withdrawal payload
func validateWalletWithdrawalPayload(payload *WalletWithdrawalPayload) error {
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


