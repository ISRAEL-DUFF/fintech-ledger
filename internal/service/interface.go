package service

import (
	"context"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
)

// TransactionService defines the interface for transaction operations
type TransactionService interface {
	// Entry operations
	CreateEntry(ctx context.Context, entry *models.Entry) error
	GetEntryByID(ctx context.Context, id string) (*models.Entry, error)
	GetEntriesByDateRange(ctx context.Context, startDate, endDate time.Time, page, pageSize int) ([]*models.Entry, int64, error)
	ValidateEntry(ctx context.Context, entry *models.Entry) error

	// Wallet operations
	ProcessTransfer(ctx context.Context, req TransferRequest) (*models.Transaction, error)
	ProcessDeposit(ctx context.Context, req DepositRequest) (*models.Transaction, error)
	ProcessWithdrawal(ctx context.Context, req WithdrawalRequest) (*models.Transaction, error)
	ProcessExchange(ctx context.Context, req ExchangeRequest) (*models.Transaction, error)
	ProcessFee(ctx context.Context, req FeeRequest) (*models.Transaction, error)
	
	// Reversal operations
	ReverseTransfer(ctx context.Context, transactionID string) error
	ReverseDeposit(ctx context.Context, transactionID string) error
	ReverseWithdrawal(ctx context.Context, transactionID string) error
	ReverseExchange(ctx context.Context, transactionID string) error
	ReverseFee(ctx context.Context, transactionID string) error
}

// TransferRequest defines the request for a transfer operation
type TransferRequest struct {
	SourceAccountID      string  `json:"source_account_id"`
	DestinationAccountID string  `json:"destination_account_id"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
	Reference            string  `json:"reference,omitempty"`
}

// DepositRequest defines the request for a deposit operation
type DepositRequest struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Reference string  `json:"reference,omitempty"`
}

// WithdrawalRequest defines the request for a withdrawal operation
type WithdrawalRequest struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Reference string  `json:"reference,omitempty"`
}

// ExchangeRequest defines the request for a currency exchange operation
type ExchangeRequest struct {
	SourceAccountID      string  `json:"source_account_id"`
	SourceAmount         float64 `json:"source_amount"`
	SourceCurrency       string  `json:"source_currency"`
	DestinationAccountID string  `json:"destination_account_id"`
	DestinationAmount    float64 `json:"destination_amount"`
	DestinationCurrency  string  `json:"destination_currency"`
	ExchangeRate         float64 `json:"exchange_rate"`
	Reference            string  `json:"reference,omitempty"`
}

// FeeRequest defines the request for a fee operation
type FeeRequest struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Reference string  `json:"reference,omitempty"`
	FeeType   string  `json:"fee_type,omitempty"`
}
