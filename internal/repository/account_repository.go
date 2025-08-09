package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/models"
	"gorm.io/gorm"
)

// AccountRepository defines the interface for account data operations.
type AccountRepository interface {
	CreateAccount(ctx context.Context, account *models.Account) error
	GetAccountByID(ctx context.Context, id string) (*models.Account, error)
	GetAccountsByUserID(ctx context.Context, userID string) ([]*models.Account, error)
	UpdateAccount(ctx context.Context, account *models.Account) error
	DeleteAccount(ctx context.Context, id string) error // Soft delete might be preferred in production
}

// accountRepository implements AccountRepository using GORM.
type accountRepository struct {
	db *gorm.DB
}

// NewAccountRepository creates a new AccountRepository.
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

// CreateAccount creates a new account in the database using GORM.
func (r *accountRepository) CreateAccount(ctx context.Context, account *models.Account) error {
	if account.ID == "" {
		account.ID = uuid.New().String()
	}
	// GORM automatically handles CreatedAt and UpdatedAt if you embed gorm.Model,
	// but since we have custom fields, we'll set them manually or rely on hooks.
	// For now, we'll ensure they are set if zero.
	if account.CreatedAt.IsZero() {
		account.CreatedAt = time.Now()
	}
	account.UpdatedAt = time.Now() // Ensure UpdatedAt is set on creation as well

	result := r.db.WithContext(ctx).Create(account)
	if result.Error != nil {
		return fmt.Errorf("failed to create account: %w", result.Error)
	}
	return nil
}

// GetAccountByID retrieves an account by its ID using GORM.
func (r *accountRepository) GetAccountByID(ctx context.Context, id string) (*models.Account, error) {
	account := &models.Account{}
	result := r.db.WithContext(ctx).First(account, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // Account not found
		}
		return nil, fmt.Errorf("failed to get account by ID: %w", result.Error)
	}
	return account, nil
}

// GetAccountsByUserID retrieves all accounts associated with a specific user ID using GORM.
func (r *accountRepository) GetAccountsByUserID(ctx context.Context, userID string) ([]*models.Account, error) {
	var accounts []*models.Account
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get accounts by user ID: %w", result.Error)
	}
	return accounts, nil
}

// UpdateAccount updates an existing account using GORM.
// Note: Immutable fields like ID, Type, Currency should not typically be changed after creation.
func (r *accountRepository) UpdateAccount(ctx context.Context, account *models.Account) error {
	account.UpdatedAt = time.Now() // Update timestamp before saving

	result := r.db.WithContext(ctx).Model(account).Where("id = ?", account.ID).Update("name", account.Name)
	if result.Error != nil {
		return fmt.Errorf("failed to update account: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("account with ID %s not found for update", account.ID)
	}
	return nil
}

// DeleteAccount deletes an account by its ID using GORM.
// In a production financial system, a 'soft delete' (marking as inactive) is almost always preferred
// over a hard delete to maintain historical integrity and audit trails.
func (r *accountRepository) DeleteAccount(ctx context.Context, id string) error {
	// For production, consider adding an 'is_active' or 'status' column and updating it to 'inactive'
	result := r.db.WithContext(ctx).Delete(&models.Account{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete account: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("account with ID %s not found for deletion", id)
	}
	return nil
}
