package models

import (
	"time"
)

// AccountType is an enumeration for different financial account types.
type AccountType string

const (
	Asset     AccountType = "Asset"
	Liability AccountType = "Liability"
	Equity    AccountType = "Equity"
	Revenue   AccountType = "Revenue"
	Expense   AccountType = "Expense"
	System    AccountType = "System" // For internal platform accounts (e.g., fees, clearing)
)

// Account represents a financial account in the ledger.
type Account struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Type      AccountType `json:"type"`
	UserID    string      `json:"user_id,omitempty"`   // Optional: For user-specific wallet accounts
	Currency  string      `json:"currency"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	// ParentAccountID string `json:"parent_account_id,omitempty"` // For hierarchical accounts
}

// EntryLine represents a single line within an Entry, affecting one account.
type EntryLine struct {
	AccountID string  `json:"account_id"`
	Debit     float64 `json:"debit"`  // Amount debited from AccountID
	Credit    float64 `json:"credit"` // Amount credited to AccountID
}

// Entry represents a single atomic financial transaction (e.g., a ledger entry).
// In double-entry bookkeeping, the sum of debits must equal the sum of credits across all lines.
type Entry struct {
	ID              string      `json:"id"`
	Description     string      `json:"description"`
	Date            time.Time   `json:"date"`
	Lines           []EntryLine `json:"lines"`
	TransactionType string      `json:"transaction_type"` // e.g., "deposit", "withdrawal", "transfer", "fee"
	ReferenceID     string      `json:"reference_id,omitempty"` // ID from an external system or parent CTE
	Status          string      `json:"status"` // e.g., "posted", "voided", "pending"
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// CTEStatus is an enumeration for the state of a Chained Transaction Event.
type CTEStatus string

const (
	CTEStatusCreated     CTEStatus = "CREATED"
	CTEStatusValidated   CTEStatus = "VALIDATED"
	CTEStatusExecuting   CTEStatus = "EXECUTING"
	CTEStatusCompleted   CTEStatus = "COMPLETED"
	CTEStatusFailed      CTEStatus = "FAILED"
	CTEStatusRollingBack CTEStatus = "ROLLING_BACK"
	CTEStatusRolledBack  CTEStatus = "ROLLED_BACK"
)

// CTETransactionStatus is an enumeration for the state of an individual transaction within a CTE.
type CTETransactionStatus string

const (
	CTETransactionStatusPending     CTETransactionStatus = "PENDING"
	CTETransactionStatusExecuting   CTETransactionStatus = "EXECUTING"
	CTETransactionStatusCompleted   CTETransactionStatus = "COMPLETED"
	CTETransactionStatusFailed      CTETransactionStatus = "FAILED"
	CTETransactionStatusCompensating CTETransactionStatus = "COMPENSATING"
	CTETransactionStatusCompensated CTETransactionStatus = "COMPENSATED"
	CTETransactionStatusSkipped     CTETransactionStatus = "SKIPPED" // For conditional transactions
)

// CompensationInfo defines the details for compensating a failed CTETransaction.
type CompensationInfo struct {
	Type            string  `json:"type"`              // e.g., "CreditWallet", "ReverseEntry"
	Amount          float64 `json:"amount,omitempty"`  // Amount for compensation if applicable
	TargetAccountID string  `json:"target_account_id,omitempty"` // Account to be affected by compensation
	Executed        bool    `json:"executed"`          // True if compensation has been performed
	EntryID         string  `json:"entry_id,omitempty"` // ID of the compensation entry
	Error           string  `json:"error,omitempty"`   // Error during compensation, if any
}

// CTETransaction represents a single step/transaction within a Chained Transaction Event.
type CTETransaction struct {
	ID                 string               `json:"id"`
	Description        string               `json:"description"`
	Type               string               `json:"type"` // e.g., "DebitWallet", "CreditMerchant", "ApplyFee"
	Status             CTETransactionStatus `json:"status"`
	Dependencies       []string             `json:"dependencies,omitempty"` // IDs of other CTETransactions it depends on
	Compensation       CompensationInfo     `json:"compensation"`           // Details for compensation
	EntryID            string               `json:"entry_id,omitempty"`     // Reference to the actual ledger Entry ID (if posted)
	ConditionalContext map[string]string    `json:"conditional_context,omitempty"` // Context for conditional execution
	Error              string               `json:"error,omitempty"`        // Error if this transaction failed
	StartedAt          *time.Time           `json:"started_at,omitempty"`
	CompletedAt        *time.Time           `json:"completed_at,omitempty"`
}

// ChainedTransactionEvent represents a high-level, multi-step transaction orchestrated by the CTE engine.
type ChainedTransactionEvent struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Status       CTEStatus        `json:"status"`
	Transactions []CTETransaction `json:"transactions"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	CompletedAt  *time.Time       `json:"completed_at,omitempty"` // When the CTE finished (success or rolled back)
	Metadata     map[string]string `json:"metadata,omitempty"`     // General purpose metadata for the event
	Error        string           `json:"error,omitempty"`        // Overall error if the CTE failed
}

// CTELienStatus is an enumeration for the state of a Chained Transaction Event Lien.
type CTELienStatus string

const (
	CTELienStatusCreated  CTELienStatus = "CREATED"
	CTELienStatusActive   CTELienStatus = "ACTIVE"
	CTELienStatusUtilized CTELienStatus = "UTILIZED"
	CTELienStatusExpired  CTELienStatus = "EXPIRED"
	CTELienStatusSettling CTELienStatus = "SETTLING"
	CTELienStatusReleased CTELienStatus = "RELEASED"
	CTELienStatusCanceled CTELienStatus = "CANCELED" // Lien explicitly canceled before use
)

// CTELien represents a lien on an account's funds within the context of a CTE.
// Funds under CTEL can be virtually utilized by transactions within the associated CTE.
type CTELien struct {
	ID          string        `json:"id"`
	CTEID       string        `json:"cte_id"`     // Link to the parent CTE
	AccountID   string        `json:"account_id"` // The account on which the lien is placed
	Amount      float64       `json:"amount"`
	Currency    string        `json:"currency"`
	Status      CTELienStatus `json:"status"`
	Description string        `json:"description,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	ExpiresAt   *time.Time    `json:"expires_at,omitempty"` // Optional: for temporary liens
	// LienSourceTransactionID string `json:"lien_source_transaction_id,omitempty"` // The transaction that created the lien
	// UtilizedByTransactionID string `json:"utilized_by_transaction_id,omitempty"` // The CTETransaction that used the lien
}
