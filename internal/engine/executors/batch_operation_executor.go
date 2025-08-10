package executors

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ISRAEL-DUFF/fintech-ledger/internal/engine/cte"
	"github.com/google/uuid"
)

// BatchOperationPayload defines the structure for batch operation payload
type BatchOperationPayload struct {
	BatchID      string                 `json:"batch_id"`
	Description  string                 `json:"description,omitempty"`
	Transactions []*BatchTransaction    `json:"transactions"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// BatchTransaction represents a single transaction within a batch
type BatchTransaction struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Payload  map[string]interface{} `json:"payload"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BatchOperationResult defines the structure for batch operation result
type BatchOperationResult struct {
	BatchID           string                    `json:"batch_id"`
	Status            string                    `json:"status"`
	ProcessedAt       time.Time                 `json:"processed_at"`
	TotalTransactions int                       `json:"total_transactions"`
	SuccessfulCount   int                       `json:"successful_count"`
	FailedCount       int                       `json:"failed_count"`
	Results           []*BatchTransactionResult `json:"results"`
}

// BatchTransactionResult represents the result of a single transaction in the batch
type BatchTransactionResult struct {
	ID        string                 `json:"id"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
	Result    map[string]interface{} `json:"result,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// BatchOperationExecutor handles batch transaction processing
type BatchOperationExecutor struct {
	executorFactory *ExecutorFactory
}

// Compensate implements the Compensate method required by the TransactionExecutor interface
// func (e *BatchOperationExecutor) Compensate(ctx context.Context, tx *cte.Transaction) error {
// 	// For batch operations, we don't implement compensation at this level
// 	// as each individual transaction in the batch should handle its own compensation
// 	return nil
// }

// NewBatchOperationExecutor creates a new BatchOperationExecutor
func NewBatchOperationExecutor(executorFactory *ExecutorFactory) *BatchOperationExecutor {
	return &BatchOperationExecutor{
		executorFactory: executorFactory,
	}
}

// Execute processes a batch of transactions
func (e *BatchOperationExecutor) Execute(ctx context.Context, tx *cte.Transaction) error {
	// Parse the batch operation payload
	payloadBytes, err := json.Marshal(tx.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var payload BatchOperationPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal batch operation payload: %w", err)
	}

	// Generate a batch ID if not provided
	if payload.BatchID == "" {
		payload.BatchID = uuid.New().String()
	}

	// Create a result structure
	result := &BatchOperationResult{
		BatchID:           payload.BatchID,
		Status:            "PROCESSING",
		ProcessedAt:       time.Now(),
		TotalTransactions: len(payload.Transactions),
		Results:           make([]*BatchTransactionResult, 0, len(payload.Transactions)),
	}

	// Process transactions in parallel with a semaphore to limit concurrency
	const maxConcurrent = 10
	sem := make(chan struct{}, maxConcurrent)
	results := make(chan *BatchTransactionResult, len(payload.Transactions))

	var wg sync.WaitGroup

	// Process each transaction in the batch
	for _, batchTx := range payload.Transactions {
		// Skip transactions without a type
		if batchTx.Type == "" {
			results <- &BatchTransactionResult{
				ID:        batchTx.ID,
				Status:    "FAILED",
				Error:     "missing transaction type",
				Timestamp: time.Now(),
			}
			continue
		}

		// Get the appropriate executor for this transaction type
		_, exists := e.executorFactory.GetExecutor(batchTx.Type)
		if !exists {
			results <- &BatchTransactionResult{
				ID:        batchTx.ID,
				Status:    "FAILED",
				Error:     fmt.Sprintf("no executor registered for type: %s", batchTx.Type),
				Timestamp: time.Now(),
			}
			continue
		}

		// Process the transaction in a goroutine
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(batchTx BatchTransaction) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			// Get the executor for this transaction type
			executor, exists := e.executorFactory.GetExecutor(batchTx.Type)
			if !exists {
				results <- &BatchTransactionResult{
					ID:        batchTx.ID,
					Status:    "FAILED",
					Error:     fmt.Sprintf("no executor registered for transaction type: %s", batchTx.Type),
					Timestamp: time.Now(),
				}
				return
			}

			// Create a transaction for the batch item
			txItem := &cte.Transaction{
				ID:        uuid.New().String(),
				EventID:   tx.EventID,
				Name:      fmt.Sprintf("Batch item: %s", batchTx.ID),
				Type:      batchTx.Type,
				Payload:   batchTx.Payload,
				State:     string(cte.TransactionStatePending),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Execute the transaction
			err := executor.Execute(ctx, txItem)

			// Record the result
			txResult := &BatchTransactionResult{
				ID:        batchTx.ID,
				Status:    "COMPLETED",
				Timestamp: time.Now(),
			}

			if err != nil {
				txResult.Status = "FAILED"
				txResult.Error = err.Error()
			} else if result, ok := txItem.Result.(map[string]interface{}); ok {
				txResult.Result = result
			}

			results <- txResult
		}(*batchTx)
	}

	// Wait for all transactions to complete and collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	for txResult := range results {
		if txResult.Status == "COMPLETED" {
			result.SuccessfulCount++
		} else {
			result.FailedCount++
		}
		result.Results = append(result.Results, txResult)
	}

	// Update the transaction result
	result.ProcessedAt = time.Now()
	if result.FailedCount == 0 {
		result.Status = "COMPLETED"
	} else if result.SuccessfulCount > 0 {
		result.Status = "PARTIALLY_COMPLETED"
	} else {
		result.Status = "FAILED"
	}

	// Update the transaction with the result
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal batch operation result: %w", err)
	}
	tx.Result = resultBytes
	tx.UpdatedAt = time.Now()

	return nil
}

// Compensate handles the rollback of a batch operation
func (e *BatchOperationExecutor) Compensate(ctx context.Context, tx *cte.Transaction) error {
	// Parse the result to get the batch operation details
	var result BatchOperationResult
	resultBytes, err := json.Marshal(tx.Result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	// If no transactions were processed, nothing to compensate
	if len(result.Results) == 0 {
		return nil
	}

	// Process compensations in parallel with a limit on concurrency
	const maxConcurrent = 5 // Use a lower concurrency for compensation

	var wg sync.WaitGroup
	errCh := make(chan error, len(result.Results))
	sem := make(chan struct{}, maxConcurrent)

	for _, txResult := range result.Results {
		// Skip transactions that didn't complete successfully
		if txResult.Status != "COMPLETED" {
			continue
		}

		// Get the transaction type from the result
		txType, ok := txResult.Result["type"].(string)
		if !ok {
			errCh <- fmt.Errorf("missing transaction type in result for transaction %s", txResult.ID)
			continue
		}

		// Get the appropriate executor for this transaction type
		executor, exists := e.executorFactory.GetExecutor(txType)
		if !exists {
			errCh <- fmt.Errorf("no executor registered for type: %s", txType)
			continue
		}

		// Create a transaction for compensation
		txItem := &cte.Transaction{
			ID:      txResult.ID,
			EventID: tx.EventID,
			Type:    txType,
			Result:  txResult.Result,
		}

		// Process compensation in a goroutine
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(txItem *cte.Transaction) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			if err := executor.Compensate(ctx, txItem); err != nil {
				errCh <- fmt.Errorf("failed to compensate transaction %s: %w", txItem.ID, err)
			}
		}(txItem)
	}

	// Close the error channel when all goroutines are done
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Collect errors
	var compensationErrors []error
	for err := range errCh {
		if err != nil {
			compensationErrors = append(compensationErrors, err)
		}
	}

	// Update the result with compensation status
	result.Status = "COMPENSATED"
	if len(compensationErrors) > 0 {
		result.Status = "PARTIALLY_COMPENSATED"
	}

	// Update the transaction result
	resultBytes, err = json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(resultBytes, &resultMap); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	tx.Result = resultMap
	tx.UpdatedAt = time.Now()

	// Return any compensation errors
	if len(compensationErrors) > 0 {
		return fmt.Errorf("compensation completed with %d errors: %v", len(compensationErrors), compensationErrors)
	}

	return nil
}

// validateBatchOperationPayload validates the batch operation payload
func validateBatchOperationPayload(payload *BatchOperationPayload) error {
	if payload == nil {
		return fmt.Errorf("payload cannot be nil")
	}

	if len(payload.Transactions) == 0 {
		return fmt.Errorf("at least one transaction is required")
	}

	// Validate each transaction in the batch
	for i, tx := range payload.Transactions {
		if tx == nil {
			return fmt.Errorf("transaction at index %d is nil", i)
		}

		if tx.Type == "" {
			return fmt.Errorf("transaction at index %d is missing type", i)
		}

		if tx.Payload == nil {
			return fmt.Errorf("transaction at index %d is missing payload", i)
		}
	}

	return nil
}
