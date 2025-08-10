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

// CurrencyExchangePayload defines the structure for currency exchange transaction payload
type CurrencyExchangePayload struct {
	SourceAccountID      string  `json:"source_account_id"`
	SourceCurrency       string  `json:"source_currency"`
	SourceAmount         float64 `json:"source_amount"`
	DestinationAccountID string  `json:"destination_account_id"`
	DestinationCurrency  string  `json:"destination_currency"`
	ExchangeRate         float64 `json:"exchange_rate"`
	Reference            string  `json:"reference,omitempty"`
	FeeAccountID         string  `json:"fee_account_id,omitempty"`
	FeeAmount            float64 `json:"fee_amount,omitempty"`
	FeeCurrency          string  `json:"fee_currency,omitempty"`
}

// CurrencyExchangeResult represents the result of a currency exchange transaction
type CurrencyExchangeResult struct {
	ID                   string    `json:"id"`
	Status               string    `json:"status"`
	SourceAccountID      string    `json:"source_account_id"`
	DestinationAccountID string    `json:"destination_account_id"`
	SourceAmount         float64   `json:"source_amount"`
	DestinationAmount    float64   `json:"destination_amount"`
	ExchangeRate         float64   `json:"exchange_rate"`
	FeeAmount            float64   `json:"fee_amount"`
	ProcessedAt          time.Time `json:"processed_at"`
	Error                string    `json:"error,omitempty"`
}

// CurrencyExchangeExecutor handles currency exchange transactions
type CurrencyExchangeExecutor struct {
	txService     service.TransactionService
	accountRepo   repository.AccountRepository
	exchangeSvc   service.ExchangeRateService
	db            *gorm.DB
	lienManager   ctel.LienManager
	transactionSvc service.TransactionService
}

// NewCurrencyExchangeExecutor creates a new currency exchange executor
func NewCurrencyExchangeExecutor(
	db *gorm.DB,
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	transactionSvc service.TransactionService,
	rateSvc service.ExchangeRateService,
	lienManager ctel.LienManager,
) *CurrencyExchangeExecutor {
	return &CurrencyExchangeExecutor{
		db:            db,
		accountRepo:   accountRepo,
		transactionSvc:  transactionSvc,
		exchangeSvc:   rateSvc,
		lienManager:   lienManager,
	}
}

// Execute performs a currency exchange between two accounts
func (e *CurrencyExchangeExecutor) Execute(ctx context.Context, tx *cte.Transaction) error {
	var err error

	// Parse the payload
	payloadBytes, err := json.Marshal(tx.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload CurrencyExchangePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Set default destination account if not provided
	if payload.DestinationAccountID == "" {
		payload.DestinationAccountID = payload.SourceAccountID
	}

	// Validate payload
	if err := validateCurrencyExchangePayload(&payload); err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	// Start a database transaction
	dbTx := e.db.Begin()
	if dbTx.Error != nil {
		return fmt.Errorf("failed to begin database transaction: %w", dbTx.Error)
	}

	// Deferred rollback in case of panic or error
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
			panic(r)
		} else if err != nil {
			dbTx.Rollback()
		}
	}()

	// Get the source account
	sourceAccount, err := e.accountRepo.GetAccountByID(ctx, payload.SourceAccountID)
	if err != nil {
		return fmt.Errorf("failed to get source account: %w", err)
	}

	// Get the destination account
	destAccount, err := e.accountRepo.GetAccountByID(ctx, payload.DestinationAccountID)
	if err != nil {
		return fmt.Errorf("failed to get destination account: %w", err)
	}

	// Check if accounts support the specified currencies
	if !accountSupportsCurrency(sourceAccount, payload.SourceCurrency) {
		return fmt.Errorf("source account does not support currency %s", payload.SourceCurrency)
	}

	if !accountSupportsCurrency(destAccount, payload.DestinationCurrency) {
		return fmt.Errorf("destination account does not support currency %s", payload.DestinationCurrency)
	}

	// Calculate destination amount
	destinationAmount := payload.SourceAmount * payload.ExchangeRate

	// Create exchange request
	exchangeReq := service.ExchangeRequest{
		SourceAccountID:      payload.SourceAccountID,
		SourceCurrency:       payload.SourceCurrency,
		SourceAmount:         payload.SourceAmount,
		DestinationAccountID: payload.DestinationAccountID,
		DestinationAmount:    destinationAmount,
		DestinationCurrency:  payload.DestinationCurrency,
		ExchangeRate:         payload.ExchangeRate,
		Reference:            payload.Reference,
	}

	// Process the exchange transaction
	if _, err := e.transactionSvc.ProcessExchange(ctx, exchangeReq); err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to process exchange: %w", err)
	}

	// Process fee if applicable
	if payload.FeeAmount > 0 {
		_, err = e.txService.ProcessFee(ctx, service.FeeRequest{
			AccountID: payload.SourceAccountID,
			Amount:    payload.FeeAmount,
			Currency:  payload.SourceCurrency,
			Reference: fmt.Sprintf("Exchange fee for %s", payload.Reference),
		})

		if err != nil {
			dbTx.Rollback()
			return fmt.Errorf("failed to process fee: %w", err)
		}
	}

	// Commit the transaction
	if err := dbTx.Commit().Error; err != nil {
		dbTx.Rollback()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Update the transaction result
	txResult := &CurrencyExchangeResult{
		ID:                   tx.ID,
		Status:               "completed",
		SourceAccountID:      payload.SourceAccountID,
		DestinationAccountID: payload.DestinationAccountID,
		SourceAmount:         payload.SourceAmount,
		DestinationAmount:    destinationAmount,
		ExchangeRate:         payload.ExchangeRate,
		FeeAmount:            payload.FeeAmount,
		ProcessedAt:          time.Now(),
	}

	// Update the transaction result
	resultJSON, err := json.Marshal(txResult)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction result: %w", err)
	}
	tx.Result = resultJSON

	// Commit the transaction
	if err := dbTx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Compensate handles the rollback of a currency exchange
func (e *CurrencyExchangeExecutor) Compensate(ctx context.Context, tx *cte.Transaction) error {
	// Parse the payload
	payloadBytes, err := json.Marshal(tx.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload CurrencyExchangePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Parse the result to get transaction IDs
	var txResult CurrencyExchangeResult
	if tx.Result != nil {
		resultBytes, err := json.Marshal(tx.Result)
		if err != nil {
			return fmt.Errorf("failed to marshal transaction result: %w", err)
		}
		if err := json.Unmarshal(resultBytes, &txResult); err != nil {
			return fmt.Errorf("failed to unmarshal transaction result: %w", err)
		}
	}

	// Start a database transaction
	dbTx := e.db.Begin()
	if dbTx.Error != nil {
		return fmt.Errorf("failed to begin database transaction: %w", dbTx.Error)
	}

	// Store the error that might occur during compensation
	var compErr error

	// Use a closure to handle the deferred rollback
	defer func() {
		if r := recover(); r != nil {
			dbTx.Rollback()
			panic(r)
		} else if compErr != nil {
			dbTx.Rollback()
		}
	}()

	// Reverse the exchange transaction if we have a transaction ID
	if txResult.ID != "" {
		// Use the transaction ID directly as per the TransactionService interface
		if err := e.transactionSvc.ReverseExchange(ctx, txResult.ID); err != nil {
			compErr = fmt.Errorf("failed to reverse exchange transaction: %w", err)
			return compErr
		}

		// Reverse the fee transaction if it exists
		if payload.FeeAmount > 0 {
			// Since we don't have the fee transaction ID in the result, we'll skip this for now
			// In a real implementation, we would store the fee transaction ID in the result
		}
	}

	// Release any liens that were placed
	if tx.ID != "" {
		// Note: This assumes the LienManager interface has a method to release all liens for an event
		// If not, you'll need to implement this functionality or adjust the compensation logic
		// For now, we'll skip this part as it depends on the LienManager implementation
		// TODO: Implement proper lien release logic when LienManager interface is updated
	}

	// Commit the transaction if no errors occurred
	if compErr == nil {
		if err := dbTx.Commit().Error; err != nil {
			compErr = fmt.Errorf("failed to commit transaction: %w", err)
			return compErr
		}
	}

	return compErr
}

// validateCurrencyExchangePayload validates the currency exchange payload
func validateCurrencyExchangePayload(payload *CurrencyExchangePayload) error {
	if payload == nil {
		return fmt.Errorf("payload is required")
	}

	if payload.SourceAccountID == "" {
		return fmt.Errorf("source_account_id is required")
	}

	if payload.SourceCurrency == "" {
		return fmt.Errorf("source_currency is required")
	}

	if payload.SourceAmount <= 0 {
		return fmt.Errorf("source_amount must be greater than 0")
	}

	if payload.DestinationAccountID == "" {
		return fmt.Errorf("destination_account_id is required")
	}

	if payload.DestinationCurrency == "" {
		return fmt.Errorf("destination_currency is required")
	}

	if payload.ExchangeRate <= 0 {
		return fmt.Errorf("exchange_rate must be greater than 0")
	}

	// If fee is provided, validate fee fields
	if payload.FeeAmount > 0 {
		if payload.FeeAccountID == "" {
			return fmt.Errorf("fee_account_id is required when fee_amount is provided")
		}
		if payload.FeeCurrency == "" {
			return fmt.Errorf("fee_currency is required when fee_amount is provided")
		}
	}

	return nil
}
