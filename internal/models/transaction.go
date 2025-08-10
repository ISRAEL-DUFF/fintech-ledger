package models

import "time"

// TransactionType represents the type of a transaction
type TransactionType string

// Transaction statuses
const (
	TransactionTypeTransfer   TransactionType = "TRANSFER"
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdrawal TransactionType = "WITHDRAWAL"
	TransactionTypeFee        TransactionType = "FEE"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

// Transaction statuses
const (
	TransactionStatusPending    TransactionStatus = "PENDING"
	TransactionStatusProcessing TransactionStatus = "PROCESSING"
	TransactionStatusCompleted  TransactionStatus = "COMPLETED"
	TransactionStatusFailed     TransactionStatus = "FAILED"
	TransactionStatusReversed   TransactionStatus = "REVERSED"
)

// Transaction represents a financial transaction in the system
type Transaction struct {
	ID              string          `json:"id" gorm:"primaryKey"`
	Type            TransactionType `json:"type" gorm:"type:varchar(20);not null;index"`
	Status          TransactionStatus `json:"status" gorm:"type:varchar(20);not null;index"`
	SourceAccountID string          `json:"source_account_id,omitempty" gorm:"index"`
	TargetAccountID string          `json:"target_account_id,omitempty" gorm:"index"`
	Amount          float64         `json:"amount" gorm:"type:decimal(19,4);not null"`
	Currency        string          `json:"currency" gorm:"type:varchar(3);not null"`
	Fee             float64         `json:ee,omitempty" gorm:"type:decimal(19,4);default:0"`
	FeeCurrency     string          `json:"fee_currency,omitempty" gorm:"type:varchar(3)"`
	Reference       string          `json:"reference,omitempty" gorm:"type:varchar(255)"`
	Description     string          `json:"description,omitempty" gorm:"type:text"`
	Metadata        JSONMap         `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt       time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
}

// JSONMap is a map that can be stored as JSON in the database
type JSONMap map[string]interface{}

// TableName specifies the table name for the Transaction model
func (Transaction) TableName() string {
	return "transactions"
}
