# Stage 3: Entry and Transaction Management Tasks

This document outlines the detailed tasks for implementing entry and transaction management for the fintech wallet ledger system. With the introduction of the CTE engine, the process of creating and managing entries will be orchestrated by the CTE engine, ensuring complex multi-transaction operations are handled atomically and consistently.

## Tasks for Core Entry/Transaction Management (to be orchestrated by CTE Engine)

### 3.1 Define and Implement Entry and EntryLine Data Structures

*   Define and implement the Go structs for `Entry` and `EntryLine`.
*   Ensure `Entry` includes fields for `ID`, `Description`, `Date`, `TransactionType`, `Source`, and a slice of `EntryLine`.
*   Ensure `EntryLine` includes fields for `AccountID`, `Debit`, and `Credit`.
*   Add appropriate JSON tags to the struct fields for API serialization.
*   Consider any additional metadata fields required for tracking transaction details (e.g., external reference IDs).

### 3.2 Design Database Schema for Entries and Entry Lines

*   **Note:** The database schema should accommodate linking entries to a parent Chained Transaction Event (CTE).
*   Design the database table schema for storing `Entry` records.
    *   Include columns for all `Entry` fields.
    *   Set the `ID` as the primary key.
    *   Add appropriate indexes on frequently queried fields (e.g., `Date`, `TransactionType`, `Source`).
    *   **Add a foreign key column linking to the `ChainedTransactionEvent` table (`cte_id`) to associate entries with their parent CTE.** This column should be nullable for entries not part of a CTE (though ideally, all entries in a fintech system would be part of a CTE for better management).
*   Design the database table schema for storing `EntryLine` records.
    *   Include columns for all `EntryLine` fields.
    *   Include a foreign key column linking back to the `Entry` table (`entry_id`).
    *   Include a foreign key column linking to the `Account` table (`account_id`).
    *   Consider a composite primary key or a unique index on a combination of columns if necessary.
    *   Add indexes on `entry_id` and `account_id` for efficient querying.

### 3.3 Implement Data Access Layer (Repository) for Entries

*   **Note:** While a basic repository is needed, the primary interaction for creating/modifying entries will be through the CTE Engine's transaction execution logic.
*   Create a Go package or module for the entry repository.
*   Implement functions within the repository for:
    *   `CreateEntry(entry *Entry) error`: Persists a new entry and its associated entry lines to the database within a single database transaction.
    *   `GetEntryByID(id string) (*Entry, error)`: Retrieves a specific entry and its lines from the database.
    *   `GetEntriesByAccountID(accountID string, filter *EntryFilter) ([]*Entry, error)`: Retrieves entries related to a specific account, potentially with filtering by date range, transaction type, etc.
    *   `GetEntriesByUserID(userID string, filter *EntryFilter) ([]*Entry, error)`: Retrieves entries related to a specific user's accounts.
    *   Implement methods for querying entries based on other criteria (e.g., transaction type, date range).
*   Utilize the chosen database driver and ORM/SQL builder.
*   Ensure all write operations (creating entries) are wrapped in database transactions for atomicity and consistency.

### 3.4 Implement Entry Validation Logic

*   Create a function `ValidateEntry(entry *Entry) error` that performs the following checks. **This validation will be called by the CTE Engine before attempting to persist an entry:**
    *   **Balance Check:** Verify that the sum of all debits in the `EntryLine`s equals the sum of all credits.
    *   **Account ID Validation:** Ensure that all `AccountID`s in the `EntryLine`s are valid and correspond to existing accounts in the system.
    *   **Amount Validation:** Ensure that debit and credit amounts are non-negative.
    *   **Date Validation:** Ensure the entry date is valid and potentially within acceptable ranges.
    *   **Transaction Type Validation:** Validate that the `TransactionType` is one of the defined types.
    *   **Other Business Logic:** Implement any other specific validation rules based on the transaction type (e.g., checking for sufficient balance for withdrawals/transfers - although strict double-entry doesn't enforce this at the entry level, application logic often does before creating the entry).

### 3.5 Implement Functions for Creating Specific Transaction Types

*   **Note:** These functions will likely be part of the business application layer (Payment Service, Wallet Service, etc.) and will prepare the necessary data for the CTE Engine to create and execute the actual ledger entries. The CTE Engine is responsible for the atomic execution and persistence.
*   Create higher-level functions that prepare the data for common fintech transaction types to be passed to the CTE Engine. These functions define the intent of the transaction. Examples include:
    *   `CreateDeposit(userID string, accountID string, amount float64, source string) (*Entry, error)`: Creates an entry to credit a user's wallet account and debit a system holding account.
    *   `CreateWithdrawal(userID string, accountID string, amount float66, destination string) (*Entry, error)`: Creates an entry to debit a user's wallet account and credit a system holding account.
    *   `CreateTransfer(fromUserID string, fromAccountID string, toUserID string, toAccountID string, amount float64) (*Entry, error)`: Creates an entry to debit the sender's wallet account and credit the receiver's wallet account.
    *   `ApplyFee(userID string, accountID string, amount float64, feeType string) (*Entry, error)`: Creates an entry to debit a user's wallet account and credit a system revenue account.
*   These functions should handle identifying the correct system accounts involved in each transaction type.
*   Ensure these functions perform necessary checks *before* creating the entry (e.g., checking if the user has sufficient balance for withdrawals or transfers if required by business logic).

## Tasks for Ensuring Entry Immutability and Data Integrity (managed by CTE Engine and Ledger)

*   Design the system to prevent modification or deletion of entries once they have been successfully created and posted to the ledger.
*   Any corrections or reversals must be handled by creating new, offsetting entries.
*   The repository layer should not expose functions for updating or deleting entries directly.

### 3.7 Implement Error Handling for Transaction Processing

*   Define a clear set of error types for transaction processing (e.g., `ErrInsufficientFunds`, `ErrInvalidAccount`, `ErrUnbalancedEntry`). **These errors should be designed to be consumable by the CTE Engine for rollback and compensation.**
*   Implement comprehensive error handling within the entry creation and validation functions.
*   Ensure that database transaction failures are handled gracefully and result in appropriate error responses.

### 3.8 Write Unit and Integration Tests

*   Write unit tests for the entry validation logic to cover various scenarios (balanced/unbalanced, valid/invalid accounts, etc.).
*   Write unit tests for the entry repository functions (creating, retrieving entries).
*   Write integration tests to verify the end-to-end flow of creating different transaction types, including database interaction and validation.
*   Test concurrency scenarios to ensure that concurrent entry creation does not lead to data inconsistencies.
