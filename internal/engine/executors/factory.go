package executors

import (
	"context"
	"fmt"
	"sync"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/engine/cte"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/engine/ctel"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/repository"
	"github.com/ISRAEL-DUFF/fintech-ledger/internal/service"
	"gorm.io/gorm"
)

// ExecutorFactory manages the creation and retrieval of transaction executors
type ExecutorFactory struct {
	db              *gorm.DB
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
	transactionSvc  service.TransactionService
	lienManager     ctel.LienManager
	executors       map[string]cte.TransactionExecutor
	mu             sync.RWMutex
}

// NewExecutorFactory creates a new executor factory
func NewExecutorFactory(
	db *gorm.DB,
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	transactionSvc service.TransactionService,
	lienManager ctel.LienManager,
) *ExecutorFactory {
	return &ExecutorFactory{
		db:             db,
		accountRepo:    accountRepo,
		transactionRepo: transactionRepo,
		transactionSvc:  transactionSvc,
		lienManager:    lienManager,
		executors:      make(map[string]cte.TransactionExecutor),
	}
}

// RegisterExecutor registers a transaction executor for a specific transaction type
func (f *ExecutorFactory) RegisterExecutor(txType string, executor cte.TransactionExecutor) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.executors[txType] = executor
}

// GetExecutor returns the executor for the given transaction type
func (f *ExecutorFactory) GetExecutor(txType string) (cte.TransactionExecutor, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	executor, exists := f.executors[txType]
	return executor, exists
}

// GetAllExecutors returns a map of all registered executors
func (f *ExecutorFactory) GetAllExecutors() map[string]cte.TransactionExecutor {
	// Create a new map to avoid external modifications
	executors := make(map[string]cte.TransactionExecutor, len(f.executors))
	for k, v := range f.executors {
		executors[k] = v
	}
	return executors
}

// InitializeDefaultExecutors registers all default transaction executors
func (f *ExecutorFactory) InitializeDefaultExecutors(ctx context.Context) error {
	// Get exchange rate service if available
	var rateSvc service.ExchangeRateService
	if svc, ok := f.transactionSvc.(interface{ GetExchangeRateService() service.ExchangeRateService }); ok {
		rateSvc = svc.GetExchangeRateService()
	}

	// Register batch operation executor first (it will be used by other executors)
	batchOpExecutor := NewBatchOperationExecutor(f)
	f.RegisterExecutor("batch.operation", batchOpExecutor)

	// Register wallet transfer executor
	transferExecutor := NewWalletTransferExecutor(
		f.db,
		f.accountRepo,
		f.transactionSvc,
	)
	f.RegisterExecutor("wallet.transfer", transferExecutor)

	// Register wallet deposit executor
	walletDepositExecutor := NewWalletDepositExecutor(
		f.db,
		f.accountRepo,
		f.transactionSvc,
	)
	f.RegisterExecutor("wallet.deposit", walletDepositExecutor)

	// Register wallet withdrawal executor
	walletWithdrawalExecutor := NewWalletWithdrawalExecutor(
		f.db,
		f.accountRepo,
		f.transactionSvc,
		f.lienManager,
	)
	f.RegisterExecutor("wallet.withdrawal", walletWithdrawalExecutor)

	// Register currency exchange executor if rate service is available
	if rateSvc != nil {
		currencyExchangeExecutor := NewCurrencyExchangeExecutor(
			f.db,
			f.accountRepo,
			f.transactionRepo,
			f.transactionSvc,
			rateSvc,
			f.lienManager,
		)
		f.RegisterExecutor("wallet.exchange", currencyExchangeExecutor)
	}

	return nil
}

// ExecuteTransaction executes a transaction using the appropriate executor
func (f *ExecutorFactory) ExecuteTransaction(ctx context.Context, tx *cte.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	executor, exists := f.GetExecutor(tx.Type)
	if !exists {
		return fmt.Errorf("no executor registered for transaction type: %s", tx.Type)
	}

	if err := executor.Execute(ctx, tx); err != nil {
		// Log the error for debugging
		// You might want to add proper logging here
		return fmt.Errorf("failed to execute transaction %s: %w", tx.ID, err)
	}

	return nil
}

// CompensateTransaction compensates a transaction using the appropriate executor
func (f *ExecutorFactory) CompensateTransaction(ctx context.Context, tx *cte.Transaction) error {
	if tx == nil {
		return fmt.Errorf("transaction cannot be nil")
	}

	executor, exists := f.GetExecutor(tx.Type)
	if !exists {
		return fmt.Errorf("no executor registered for transaction type: %s", tx.Type)
	}

	if err := executor.Compensate(ctx, tx); err != nil {
		// Log the error for debugging
		// You might want to add proper logging here
		return fmt.Errorf("failed to compensate transaction %s: %w", tx.ID, err)
	}

	return nil
}
