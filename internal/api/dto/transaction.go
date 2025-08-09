package dto

import (
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
	"github.com/google/uuid"
)

// TransactionLineEntry represents a single transaction line entry
// swagger:model TransactionLineEntry
type TransactionLineEntry struct {
	// The unique identifier of the transaction line
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`

	// The account ID for this transaction line
	// required: true
	// example: 550e8400-e29b-41d4-a716-446655440000
	AccountID string `json:"account_id" validate:"required,uuid"`

	// The amount for this line (positive for credit, negative for debit)
	// required: true
	// example: 100.50
	Amount float64 `json:"amount" validate:"required,numeric"`

	// The currency code (ISO 4217)
	// required: true
	// example: USD
	Currency string `json:"currency" validate:"required,iso4217"`

	// Description of the transaction line
	// example: Payment for services
	Description string `json:"description,omitempty"`

	// The date and time when the line was created
	// example: 2023-01-01T00:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// TransactionRequest represents a transaction creation request
// swagger:model TransactionRequest
type TransactionRequest struct {
	// The transaction reference ID (optional, will be generated if not provided)
	// example: TXN-12345
	Reference string `json:"reference,omitempty"`

	// The transaction description
	// required: true
	// example: Payment for services
	Description string `json:"description" validate:"required"`

	// The date of the transaction (ISO 8601 format)
	// example: 2023-01-01T00:00:00Z
	TransactionDate time.Time `json:"transaction_date,omitempty"`

	// The transaction lines (debits and credits)
	// required: true
	// minItems: 2
	Lines []TransactionLineEntry `json:"lines" validate:"required,min=2,dive"`
}

// TransactionLineResponse represents a single transaction line in a response
// swagger:model TransactionLineResponse
type TransactionLineResponse struct {
	// The unique identifier of the transaction line
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`

	// The account ID for this transaction line
	// example: 550e8400-e29b-41d4-a716-446655440001
	AccountID string `json:"account_id"`

	// The amount for this line
	// example: 100.50
	Amount float64 `json:"amount"`

	// The currency code (ISO 4217)
	// example: USD
	Currency string `json:"currency"`

	// Description of the transaction line
	// example: Payment for services
	Description string `json:"description,omitempty"`

	// The date and time when the line was created
	// example: 2023-01-01T00:00:00Z
	CreatedAt time.Time `json:"created_at"`
}

// TransactionResponse represents a transaction in the API response
// swagger:model TransactionResponse
type TransactionResponse struct {
	// The unique identifier of the transaction
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`

	// The transaction reference ID
	// example: TXN-12345
	Reference string `json:"reference"`

	// The transaction description
	// example: Payment for services
	Description string `json:"description"`

	// The transaction status
	// example: posted
	Status string `json:"status"`

	// The type of transaction (debit/credit)
	// example: debit
	TransactionType string `json:"transaction_type"`

	// The reference ID for the transaction
	// example: REF-12345
	ReferenceID string `json:"reference_id"`

	// The date of the transaction
	// example: 2023-01-01T00:00:00Z
	Date time.Time `json:"date"`

	// The transaction lines (debits and credits)
	Lines []TransactionLineEntry `json:"lines"`

	// The date and time when the transaction was created
	// example: 2023-01-01T00:00:00Z
	CreatedAt time.Time `json:"created_at"`

	// The date and time when the transaction was last updated
	// example: 2023-01-01T00:00:00Z
	UpdatedAt time.Time `json:"updated_at"`
}

// TransactionsListResponse represents a paginated list of transactions
// swagger:model TransactionsListResponse
type TransactionsListResponse struct {
	// The list of transactions
	Data []TransactionResponse `json:"data"`

	// The total number of transactions
	Total int64 `json:"total"`

	// The current page number
	Page int `json:"page"`

	// The number of transactions per page
	PageSize int `json:"page_size"`

	// The total number of pages
	TotalPages int `json:"total_pages"`
}

// ErrorResponse represents an error response
// swagger:model ErrorResponse
type ErrorResponse struct {
	// The error message
	// example: An error occurred
	Error string `json:"error"`

	// The error code
	// example: 400
	Status int `json:"status"`
}

// CreateTransactionRequest represents the request payload for creating a transaction
// swagger:model CreateTransactionRequest
type CreateTransactionRequest struct {
	// Description of the transaction
	// required: true
	// max length: 255
	// example: Payment for services
	Description string `json:"description" validate:"required,max=255"`

	// Type of transaction (debit/credit)
	// required: true
	// enum: debit,credit
	// example: debit
	TransactionType string `json:"transaction_type" validate:"required,oneof=fund transfer withdraw"`

	// Reference ID for the transaction
	// example: REF-12345
	ReferenceID string `json:"reference_id"`

	// Date of the transaction
	// example: 2023-01-01T00:00:00Z
	Date time.Time `json:"date"`

	// Transaction lines (debits and credits)
	// required: true
	// min items: 2
	Lines []TransactionLineEntry `json:"lines" validate:"required,min=2,dive"`
}

// TransactionLine represents a single line in a transaction
type TransactionLine struct {
	AccountID string  `json:"account_id" validate:"required,uuid4"`
	Debit     float64 `json:"debit" validate:"gte=0"`
	Credit    float64 `json:"credit" validate:"gte=0"`
}

// ToModel converts a CreateTransactionRequest to a models.Entry
func (r CreateTransactionRequest) ToModel() *models.Entry {
	now := time.Now()
	entry := &models.Entry{
		ID:              uuid.New().String(),
		Description:     r.Description,
		TransactionType: r.TransactionType,
		ReferenceID:     r.ReferenceID,
		Date:            r.Date,
		Status:          "posted",
		CreatedAt:       now,
		UpdatedAt:       now,
		Lines:           make([]models.EntryLine, 0, len(r.Lines)),
	}

	// Convert lines
	for _, line := range r.Lines {
		entryLine := models.EntryLine{
			ID:        uuid.New().String(),
			EntryID:   entry.ID,
			AccountID: line.AccountID,
			CreatedAt: now,
		}

		// Set debit or credit based on amount sign
		if line.Amount > 0 {
			entryLine.Credit = line.Amount
		} else if line.Amount < 0 {
			entryLine.Debit = -line.Amount // Store as positive
		}

		entry.Lines = append(entry.Lines, entryLine)
	}

	return entry
}

// ToResponse converts a models.Entry to a TransactionResponse
func ToResponse(entry *models.Entry) *TransactionResponse {
	if entry == nil {
		return nil
	}

	resp := &TransactionResponse{
		ID:              entry.ID,
		Description:     entry.Description,
		TransactionType: entry.TransactionType,
		ReferenceID:     entry.ReferenceID,
		Status:          entry.Status,
		Date:            entry.Date,
		CreatedAt:       entry.CreatedAt,
		UpdatedAt:       entry.UpdatedAt,
		Lines:           make([]TransactionLineEntry, 0, len(entry.Lines)),
	}

	for _, line := range entry.Lines {
		lineEntry := TransactionLineEntry{
			ID:        line.ID,
			AccountID: line.AccountID,
			CreatedAt: line.CreatedAt,
		}

		// Set amount based on debit/credit
		if line.Debit > 0 {
			lineEntry.Amount = -line.Debit // Negative for debits
		} else if line.Credit > 0 {
			lineEntry.Amount = line.Credit // Positive for credits
		}

		resp.Lines = append(resp.Lines, lineEntry)
	}

	return resp
}
