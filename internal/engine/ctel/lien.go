package ctel

import (
	"context"
	"time"
)

// LienState represents the state of a CTEL lien
type LienState string

const (
	// LienStatePending indicates the lien is pending activation
	LienStatePending LienState = "PENDING"
	// LienStateActive indicates the lien is active and funds are reserved
	LienStateActive LienState = "ACTIVE"
	// LienStateReleased indicates the lien has been released
	LienStateReleased LienState = "RELEASED"
	// LienStateExpired indicates the lien has expired
	LienStateExpired LienState = "EXPIRED"
)

// Lien represents a Chained Transaction Event Lien
// This entity tracks funds that are reserved for a specific CTE event
// and can be used for transaction execution within that event.
type Lien struct {
	// ID is the unique identifier for the lien
	ID string `json:"id"`
	// EventID is the ID of the CTE event this lien is associated with
	EventID string `json:"event_id"`
	// AccountID is the ID of the account the lien is placed on
	AccountID string `json:"account_id"`
	// Amount is the amount of funds reserved by the lien
	Amount float64 `json:"amount"`
	// Currency is the currency of the lien amount
	Currency string `json:"currency"`
	// State is the current state of the lien
	State LienState `json:"state"`
	// ExpiresAt is the timestamp when the lien expires
	ExpiresAt time.Time `json:"expires_at"`
	// CreatedAt is the timestamp when the lien was created
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is the timestamp when the lien was last updated
	UpdatedAt time.Time `json:"updated_at"`
	// Metadata contains additional context or parameters for the lien
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// LienManager manages the lifecycle of CTE-liens
type ILienManager interface {
	// CreateLien creates a new lien for a CTE event
	CreateLien(
		ctx context.Context,
		eventID string,
		accountID string,
		amount float64,
		currency string,
		expiresAt time.Time,
		metadata map[string]interface{},
	) (*Lien, error)

	// GetLien retrieves a lien by ID
	GetLien(ctx context.Context, id string) (*Lien, error)

	// GetLiensByEvent retrieves all liens for a specific CTE event
	GetLiensByEvent(ctx context.Context, eventID string) ([]*Lien, error)

	// GetLiensByAccount retrieves all liens for a specific account
	GetLiensByAccount(ctx context.Context, accountID string) ([]*Lien, error)

	// ActivateLien activates a pending lien
	ActivateLien(ctx context.Context, id string) error

	// ReleaseLien releases an active lien
	ReleaseLien(ctx context.Context, id string) error

	// ExpireLien marks an expired lien as expired
	ExpireLien(ctx context.Context, id string) error

	// GetAvailableBalance calculates the available balance for an account,
	// taking into account active liens within the context of a CTE event
	GetAvailableBalance(
		ctx context.Context,
		eventID string,
		accountID string,
	) (float64, error)
}

// LienStore persists the state of CTE-liens
type LienStore interface {
	// SaveLien saves a lien to the store
	SaveLien(ctx context.Context, lien *Lien) error
	// GetLien retrieves a lien by ID
	GetLien(ctx context.Context, id string) (*Lien, error)
	// GetLiensByEvent retrieves all liens for a specific CTE event
	GetLiensByEvent(ctx context.Context, eventID string) ([]*Lien, error)
	// GetLiensByAccount retrieves all liens for a specific account
	GetLiensByAccount(ctx context.Context, accountID string) ([]*Lien, error)
	// UpdateLien updates an existing lien
	UpdateLien(ctx context.Context, lien *Lien) error
}
