# Stage 1: Account Management

This document outlines the plan for implementing the account management components of the double-entry financial ledger system.

## 1. Data Structures and Database Schema

We will define the core data structures for representing accounts.

### `Account` Struct

```
go
type Account struct {
	ID       string `json:"id"` // Unique identifier for the account (e.g., UUID)
	Name     string `json:"name"`
	Type     AccountType `json:"type"` // e.g., Asset, Liability, Equity, Revenue, Expense
	ParentID *string `json:"parent_id"` // Optional: for hierarchical account structures
	// Add fields for metadata like creation date, last updated date, etc.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```
### `AccountType` Enum
```
go
type AccountType string

const (
	Asset    AccountType = "Asset"
	Liability AccountType = "Liability"
	Equity    AccountType = "Equity"
	Revenue   AccountType = "Revenue"
	Expense   AccountType = "Expense"
)
```
### Storage Considerations

We will need a persistent storage mechanism for accounts. Options include:

*   **Relational Database (e.g., PostgreSQL, MySQL):** Provides structured storage, ACID properties, and good querying capabilities.
*   **NoSQL Database (e.g., MongoDB, BoltDB):** More flexible schema, potentially faster for writes depending on the structure.
*   **File-based Storage (e.g., JSON, CSV):** Simpler for small systems or prototyping, but less performant and scalable.

For a robust system, a relational database is recommended.

## 2. Account Management Functions

We will implement a set of functions to manage accounts. These functions will interact with the chosen storage layer.

### `CreateAccount(account Account) error`

*   **Description:** Creates a new account in the system.
*   **Input:** An `Account` struct.
*   **Validation:**
    *   Ensure `ID` is unique.
    *   Validate `Name` is not empty.
    *   Validate `Type` is a valid `AccountType`.
    *   If `ParentID` is provided, ensure the parent account exists.
*   **Process:** Store the account data in the persistent storage.
*   **Output:** An error if creation fails.

### `GetAccountByID(id string) (*Account, error)`

*   **Description:** Retrieves an account by its unique identifier.
*   **Input:** Account ID string.
*   **Process:** Query the storage for the account with the given ID.
*   **Output:** A pointer to the `Account` struct if found, or an error if not found or a storage error occurs.

### `UpdateAccount(account Account) error`

*   **Description:** Updates an existing account's details.
*   **Input:** An `Account` struct with the updated information.
*   **Validation:**
    *   Ensure the account with the given `ID` exists.
    *   Validate updated fields (e.g., `Name`, `Type`, `ParentID`).
    *   Prevent changing the `ID`.
*   **Process:** Update the account data in the persistent storage.
*   **Output:** An error if the update fails.

### `DeleteAccount(id string) error`

*   **Description:** Deletes an account from the system.
*   **Input:** Account ID string.
*   **Validation:**
    *   Ensure the account with the given `ID` exists.
    *   **Crucial:** Implement checks to prevent deleting accounts that have associated transactions. Consider archiving or marking as inactive instead of true deletion if historical data is important.
*   **Process:** Remove the account from the persistent storage (or mark as inactive).
*   **Output:** An error if deletion fails.

### `GetAllAccounts() ([]Account, error)`

*   **Description:** Retrieves all accounts in the system.
*   **Process:** Query the storage for all accounts.
*   **Output:** A slice of `Account` structs, or an error if a storage error occurs.

### `GetAccountsByType(accountType AccountType) ([]Account, error)`

*   **Description:** Retrieves accounts of a specific type.
*   **Input:** `AccountType`.
*   **Process:** Query the storage for accounts with the given type.
*   **Output:** A slice of `Account` structs, or an error.

## 3. Account Types and Hierarchies

The `AccountType` enum provides a fundamental categorization of accounts. We will need to define how these types behave in the context of debit and credit entries (e.g., assets increase with debits, liabilities increase with credits).

Implementing a hierarchical structure using `ParentID` allows for organizing accounts into categories and subcategories (e.g., "Current Assets" under "Assets").

### Considerations for Hierarchies:

*   **Recursive Queries:** Functions to retrieve all child accounts of a parent will be needed.
*   **Validation:** Prevent creating circular dependencies in the hierarchy.
*   **Reporting:** Reports may need to aggregate balances based on the hierarchy.

## 4. Implementation Details

*   **API Design:** Define the API endpoints or function signatures for accessing the account management functions.
*   **Error Handling:** Implement robust error handling for all operations, including storage errors and validation errors.
*   **Testing:** Write comprehensive unit and integration tests for all account management functions.
*   **Concurrency:** If the system needs to handle multiple users or processes concurrently, consider mechanisms for concurrency control to prevent data corruption.

This plan provides a foundation for implementing the account management stage of the ledger system. The next stage will focus on managing transactions (entries and entry lines).