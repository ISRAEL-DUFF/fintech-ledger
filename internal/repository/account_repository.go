package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"wallet-system/internal/models"
)

// AccountRepository defines the interface for account data operations.
type AccountRepository interface {
	CreateAccount(ctx context.Context, account *models.Account) error
	GetAccountByID(ctx context.Context, id string) (*models.Account, error)
	GetAccountsByUserID(ctx context.Context, userID string) ([]*models.Account, error)
	UpdateAccount(ctx context.Context, account *models.Account) error
	DeleteAccount(ctx context.Context, id string) error // Soft delete might be preferred in production
}

// accountRepository implements AccountRepository using SQL database.
type accountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates a new AccountRepository.
func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{db: db}
}

// CreateAccount creates a new account in the database.
func (r *accountRepository) CreateAccount(ctx context.Context, account *models.Account) error {
	if account.ID == "" {
		account.ID = uuid.New().String()
	}
	if account.CreatedAt.IsZero() {
		account.CreatedAt = time.Now()
	}
	account.UpdatedAt = time.Now()

	query := `INSERT INTO accounts (id, name, type, user_id, currency, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, account.ID, account.Name, account.Type, account.UserID, account.Currency, account.CreatedAt, account.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

// GetAccountByID retrieves an account by its ID.
func (r *accountRepository) GetAccountByID(ctx context.Context, id string) (*models.Account, error) {
	account := &models.Account{}
	query := `SELECT id, name, type, user_id, currency, created_at, updated_at FROM accounts WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&account.ID, &account.Name, &account.Type, &account.UserID, &account.Currency, &account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Account not found
		}
		return nil, fmt.Errorf("failed to get account by ID: %w", err)
	}
	return account, nil
}

// GetAccountsByUserID retrieves all accounts associated with a specific user ID.
func (r *accountRepository) GetAccountsByUserID(ctx context.Context, userID string) ([]*models.Account, error) {
	var accounts []*models.Account
	query := `SELECT id, name, type, user_id, currency, created_at, updated_at FROM accounts WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by user ID: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		account := &models.Account{}
		err := rows.Scan(&account.ID, &account.Name, &account.Type, &account.UserID, &account.Currency, &account.CreatedAt, &account.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account row: %w", err)
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after iterating over account rows: %w", err)
	}

	return accounts, nil
}

// UpdateAccount updates an existing account. Note: Immutable fields like ID, Type, Currency should not typically be changed after creation.
func (r *accountRepository) UpdateAccount(ctx context.Context, account *models.Account) error {
	account.UpdatedAt = time.Now()
	// Only allow updating mutable fields like Name. Type, UserID, Currency are typically immutable.
	query := `UPDATE accounts SET name = $1, updated_at = $2 WHERE id = $3`
	result, err := r.db.ExecContext(ctx, query, account.Name, account.UpdatedAt, account.ID)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("account with ID %s not found for update", account.ID)
	}
	return nil
}

// DeleteAccount deletes an account by its ID.
// In a production financial system, a 'soft delete' (marking as inactive) is almost always preferred
// over a hard delete to maintain historical integrity and audit trails.
func (r *accountRepository) DeleteAccount(ctx context.Context, id string) error {
	// For production, consider adding an 'is_active' or 'status' column and updating it to 'inactive'
	query := `DELETE FROM accounts WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("account with ID %s not found for deletion", id)
	}
	return nil
}
