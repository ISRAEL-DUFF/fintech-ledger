# CTE-CTEL Engine

The Chained Transaction Events (CTE) and Chained Transaction Event Lien (CTEL) Engine is a powerful framework for orchestrating complex, multi-step financial workflows within the fintech wallet ledger system. It provides a robust solution for managing distributed transactions, ensuring consistency, and handling failures gracefully.

## Core Concepts

### Chained Transaction Events (CTE)

CTE allows you to define a sequence of transactions that must be executed as an atomic unit. If any transaction in the sequence fails, the engine can automatically compensate for completed transactions to maintain consistency.

### Chained Transaction Event Liens (CTEL)

CTEL extends CTE by introducing the concept of liens, which are reservations of funds that can be used within the context of a specific CTE. This allows for better liquidity management and prevents double-spending across different transactions.

## Architecture

```
+------------------+       +------------------+       +------------------+
|                  |       |                  |       |                  |
|   Business       |       |      CTE         |       |    Ledger        |
|   Application    |<----->|      Engine      |<----->|    System        |
|                  |       |                  |       |                  |
+------------------+       +------------------+       +------------------+
                                    |
                                    |
                           +------------------+
                           |                  |
                           |     CTEL         |
                           |     Manager      |
                           |                  |
                           +------------------+
```

## Key Components

### Event Coordinator
Manages the lifecycle of chained transaction events, including creation, validation, and execution.

### Transaction Executor
Responsible for executing individual transactions within an event, with support for retries and error handling.

### Rollback Manager
Handles compensation logic for failed transactions, ensuring that completed transactions can be properly rolled back.

### State Manager
Tracks the state of each CTE event and its transactions, enabling recovery in case of failures.

### Lien Manager
Manages the creation, activation, and release of liens for CTE events, ensuring proper fund reservation and release.

## Usage

### Creating a CTE Event

```go
event, err := cteEngine.CreateEvent(ctx, "wallet-transfer", "Transfer between wallets", 5*time.Minute, nil)
if err != nil {
    return fmt.Errorf("failed to create event: %w", err)
}
```

### Adding Transactions to an Event

```go
tx := &cte.Transaction{
    Name:        "debit-source",
    Description: "Debit source wallet",
    Type:        "wallet.debit",
    Order:       1,
    Payload: map[string]interface{}{
        "account_id": sourceAccountID,
        "amount":     amount,
        "currency":   currency,
    },
}

if err := cteEngine.AddTransaction(ctx, event.ID, tx); err != nil {
    return fmt.Errorf("failed to add transaction: %w", err)
}
```

### Starting an Event

```go
if err := cteEngine.StartEvent(ctx, event.ID); err != nil {
    return fmt.Errorf("failed to start event: %w", err)
}
```

### Creating a Lien

```go
lien, err := lienManager.CreateLien(
    ctx,
    eventID,
    accountID,
    amount,
    currency,
    time.Now().Add(30*time.Minute),
    nil,
)
if err != nil {
    return fmt.Errorf("failed to create lien: %w", err)
}
```

## Error Handling

The CTE-CTEL engine provides comprehensive error handling and recovery mechanisms:

1. **Automatic Retries**: Failed transactions are automatically retried according to the configured retry policy.
2. **Compensation**: If a transaction fails, the engine will execute compensation logic for previously completed transactions.
3. **State Persistence**: The state of all events and transactions is persisted, allowing for recovery after restarts.

## Best Practices

1. **Idempotency**: Ensure that all transaction and compensation logic is idempotent to handle retries correctly.
2. **Timeouts**: Set appropriate timeouts for events to prevent them from running indefinitely.
3. **Monitoring**: Monitor the state of long-running events and implement alerts for stuck or failed events.
4. **Testing**: Thoroughly test all transaction and compensation logic in a staging environment before deploying to production.

## Database Schema

The CTE-CTEL engine uses the following database tables:

- `cte_events`: Stores CTE event metadata and state.
- `cte_transactions`: Stores individual transactions within CTE events.
- `cte_liens`: Tracks fund reservations for CTE events.

Refer to the migration files for the complete schema definition.

## Built-in Transaction Executors

The CTE-CTEL engine comes with several built-in transaction executors for common financial operations. These executors handle the core functionality of the wallet system and can be used as building blocks for more complex workflows.

### 1. Wallet Transfer Executor

Handles transfers between two wallet accounts.

**Transaction Type:** `wallet.transfer`

**Payload:**
```json
{
  "source_account_id": "account-123",
  "destination_account_id": "account-456",
  "amount": 100.50,
  "currency": "USD",
  "reference": "tx-ref-789",
  "description": "P2P transfer"
}
```

**Features:**
- Validates account existence and currency support
- Ensures sufficient balance in the source account
- Atomic transfer between accounts
- Comprehensive error handling

### 2. Wallet Deposit Executor

Handles deposits into wallet accounts.

**Transaction Type:** `wallet.deposit`

**Payload:**
```json
{
  "account_id": "account-123",
  "amount": 500.00,
  "currency": "USD",
  "reference": "deposit-ref-456",
  "description": "Bank transfer deposit"
}
```

**Features:**
- Validates account existence and currency support
- Supports external reference tracking
- Atomic credit operation

### 3. Wallet Withdrawal Executor

Handles withdrawals from wallet accounts with lien support.

**Transaction Type:** `wallet.withdrawal`

**Payload:**
```json
{
  "account_id": "account-123",
  "amount": 200.00,
  "currency": "USD",
  "reference": "withdrawal-ref-789",
  "description": "Withdrawal to bank"
}
```

**Features:**
- Creates a lien to reserve funds
- Validates account balance and currency support
- Supports automatic lien release on failure
- Atomic debit operation

### 4. Currency Exchange Executor

Handles currency conversions between accounts.

**Transaction Type:** `wallet.exchange`

**Payload:**
```json
{
  "source_account_id": "account-123",
  "source_currency": "USD",
  "source_amount": 100.00,
  "destination_account_id": "account-456",
  "destination_currency": "EUR",
  "exchange_rate": 0.85,
  "reference": "exchange-ref-123",
  "fee_account_id": "fees-account",
  "fee_amount": 2.50,
  "fee_currency": "USD"
}
```

**Features:**
- Handles multi-currency conversions
- Supports dynamic exchange rates
- Optional fee processing
- Atomic exchange operation
- Lien-based fund reservation

### 5. Batch Operation Executor

Processes multiple transactions as a single atomic unit.

**Transaction Type:** `batch.operation`

**Payload:**
```json
{
  "batch_id": "batch-123",
  "description": "End-of-day settlement",
  "transactions": [
    {
      "id": "tx-1",
      "type": "wallet.transfer",
      "payload": {
        "source_account_id": "account-123",
        "destination_account_id": "account-456",
        "amount": 100.00,
        "currency": "USD"
      }
    },
    {
      "id": "tx-2",
      "type": "wallet.exchange",
      "payload": {
        "source_account_id": "account-456",
        "source_currency": "USD",
        "source_amount": 50.00,
        "destination_account_id": "account-456",
        "destination_currency": "EUR",
        "exchange_rate": 0.85
      }
    }
  ]
}
```

**Features:**
- Processes multiple transactions atomically
- Supports mixed transaction types
- Configurable concurrency
- Detailed per-transaction results
- Automatic compensation on failure
- Progress tracking

## Extending the Engine

To add support for new transaction types, implement the `TransactionExecutor` interface and register it with the engine:

```go
type MyTransactionExecutor struct{}

func (e *MyTransactionExecutor) Execute(ctx context.Context, tx *cte.Transaction) error {
    // Implementation for executing the transaction
    return nil
}

func (e *MyTransactionExecutor) Compensate(ctx context.Context, tx *cte.Transaction) error {
    // Implementation for compensating the transaction
    return nil
}

// Register the executor with the factory
executorFactory.RegisterExecutor("my-transaction-type", &MyTransactionExecutor{})
```

### Example: Custom Transaction Executor

Here's an example of a custom executor for processing fee collections:

```go
type FeeCollectionExecutor struct {
    db                *gorm.DB
    transactionSvc    service.TransactionService
}

func (e *FeeCollectionExecutor) Execute(ctx context.Context, tx *cte.Transaction) error {
    // Parse and validate payload
    var payload struct {
        AccountID string  `json:"account_id"`
        Amount    float64 `json:"amount"`
        Currency  string  `json:"currency"`
        FeeType   string  `json:"fee_type"`
    }
    
    if err := mapstructure.Decode(tx.Payload, &payload); err != nil {
        return fmt.Errorf("invalid payload: %w", err)
    }
    
    // Process fee collection
    // ...
    
    // Update transaction result
    tx.Result = map[string]interface{}{
        "status":      "completed",
        "processed_at": time.Now(),
        "fee_type":    payload.FeeType,
    }
    
    return nil
}

func (e *FeeCollectionExecutor) Compensate(ctx context.Context, tx *cte.Transaction) error {
    // Implement compensation logic
    // ...
    return nil
}

// Register the executor during initialization
func init() {
    executorFactory.RegisterExecutor("fee.collection", &FeeCollectionExecutor{db, transactionSvc})
}
```
