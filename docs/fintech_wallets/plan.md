# Detailed Plan: Fintech Wallet Ledger System

This document outlines a detailed plan for implementing a robust wallet ledger system for a fintech, built upon the principles of a double-entry accounting ledger.

## 1. Mapping Fintech Concepts to Ledger Entities

We will map core fintech concepts to the fundamental entities of a double-entry ledger as follows:

*   **Wallet:** Each user's wallet will be represented as one or more `Account` entities within the ledger.
*   **Financial Transaction:** Any movement of value (money) within or across wallets will be recorded as an `Entry`.
*   **Transaction Legs:** The individual debits and credits that make up a transaction will be represented as `EntryLine` entities.

Additionally, we will introduce **System Accounts** to track funds held by the platform, revenue generated, expenses, and funds in transit.

## 2. Refined Data Structures

Building upon the basic ledger data structures, we will refine them to include fields specific to a fintech wallet system.

### Account

Represents a wallet or a system account.

```go
type Account struct {
    ID         string    `json:"id"`         // Unique identifier for the account (e.g., UUID)
    Name       string    `json:"name"`       // Human-readable name (e.g., "User X's Primary Wallet")
    Type       AccountType `json:"type"`     // e.g., Asset, Liability, Revenue, Expense
    UserID     string    `json:"user_id"`    // ID of the user if this is a user wallet account (nullable for system accounts)
    Currency   string    `json:"currency"`   // Currency of the account (e.g., "USD", "EUR")
    Status     string    `json:"status"`     // e.g., "active", "suspended", "closed"
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

type AccountType string

const (
    Asset     AccountType = "Asset"     // Represents assets the platform holds (e.g., bank accounts, funds in transit)
    Liability AccountType = "Liability" // Represents funds the platform owes to users (user wallets)
    Equity    AccountType = "Equity"    // Owner's stake (less relevant for operational ledger)
    Revenue   AccountType = "Revenue"   // Income generated (e.g., fees)
    Expense   AccountType = "Expense"   // Costs incurred
)
```

**Considerations:**

*   Choosing appropriate `AccountType`s is crucial for accurate reporting. User wallets are typically `Liability` for the platform.
*   The `UserID` links a wallet account to a specific user in the platform's user management system.
*   `Currency` ensures that transactions are recorded in the correct denomination and allows for multi-currency support.

### Entry

Represents a single financial transaction with its associated details and lines.

```go
type Entry struct {
    ID              string    `json:"id"`              // Unique identifier for the entry (transaction ID)
    Description     string    `json:"description"`     // Human-readable description of the transaction
    Date            time.Time `json:"date"`            // Date and time of the transaction
    TransactionType string    `json:"transaction_type"`// e.g., "deposit", "withdrawal", "transfer", "fee"
    ReferenceID     string    `json:"reference_id"`    // Optional: Link to an external system ID (e.g., payment gateway ID)
    Metadata        map[string]interface{} `json:"metadata"` // Optional: Store additional transaction details
    Lines           []EntryLine `json:"lines"`           // The debit and credit lines for this entry
    CreatedAt       time.Time `json:"created_at"`
}
```

**Considerations:**

*   `TransactionType` helps categorize entries for reporting and processing logic.
*   `ReferenceID` is vital for linking ledger entries back to initiating events or external system records.
*   `Metadata` provides flexibility to store transaction-specific information.

### EntryLine

Represents a single debit or credit line within an Entry, affecting a specific Account.

```go
type EntryLine struct {
    ID        string  `json:"id"`        // Unique identifier for the entry line
    EntryID   string  `json:"entry_id"`  // Foreign key linking to the parent Entry
    AccountID string  `json:"account_id"`// Foreign key linking to the affected Account
    Debit     float64 `json:"debit"`     // Debit amount (positive value)
    Credit    float64 `json:"credit"`    // Credit amount (positive value)
    // Note: Either Debit or Credit should be non-zero, not both.
}
```

**Considerations:**

*   Ensuring that for each `EntryLine`, either `Debit` or `Credit` is zero is a crucial validation rule.

## 3. Defining Transaction Types and Double-Entry Representation

Each transaction type will have a specific pattern of debits and credits to maintain the double-entry principle (Total Debits = Total Credits for each Entry).

*   **Deposit:**
    *   Debit: A system **Asset** account (e.g., "Bank Clearing Account") - Represents money entering the system.
    *   Credit: User's Wallet **Liability** account - Represents money added to the user's wallet.
    *   Example: User deposits $100. Debit "Bank Clearing Account" $100, Credit "User X Wallet" $100.

*   **Withdrawal:**
    *   Debit: User's Wallet **Liability** account - Represents money leaving the user's wallet.
    *   Credit: A system **Asset** account (e.g., "Bank Clearing Account") - Represents money leaving the system.
    *   Example: User withdraws $50. Debit "User X Wallet" $50, Credit "Bank Clearing Account" $50.

*   **Transfer (User A to User B):**
    *   Debit: User A's Wallet **Liability** account.
    *   Credit: User B's Wallet **Liability** account.
    *   Example: User A transfers $20 to User B. Debit "User A Wallet" $20, Credit "User B Wallet" $20.

*   **Fee Deduction:**
    *   Debit: User's Wallet **Liability** account - Represents money deducted from the user's wallet.
    *   Credit: A system **Revenue** account (e.g., "Transaction Fee Revenue") - Represents income for the platform.
    *   Example: Deduct $1 fee. Debit "User X Wallet" $1, Credit "Transaction Fee Revenue" $1.

*   **Payment to Merchant:**
    *   Debit: User's Wallet **Liability** account.
    *   Credit: A system **Liability** account representing the merchant's balance (or a clearing account if settling later).
    *   Example: User pays Merchant Y $30. Debit "User X Wallet" $30, Credit "Merchant Y Liability Account" $30.

**Implementation Plan:**

*   Define an enum or constant list for all supported `TransactionType`s.
*   Create functions or methods for each transaction type that generate the correct set of `EntryLine`s, ensuring debits and credits are balanced.

## 4. Implementing Wallet Operations

We will provide API endpoints and corresponding internal logic for key wallet operations. Each operation will involve creating and posting a valid `Entry`.

*   **Deposit:**
    *   API Endpoint: `POST /wallets/{user_id}/deposit`
    *   Input: `user_id`, `amount`, `currency`, `source_details` (e.g., payment gateway info).
    *   Logic:
        *   Validate input and user/wallet existence.
        *   Initiate interaction with a payment gateway to receive funds.
        *   Upon successful fund reception (via webhook or API call), create a "deposit" `Entry` with two `EntryLine`s: Debit System Asset Account, Credit User Wallet Account.
        *   Persist the `Entry` within a database transaction.
        *   Handle potential failures (e.g., payment gateway error) by not creating the entry or creating a reversal entry if partially processed.

*   **Withdrawal:**
    *   API Endpoint: `POST /wallets/{user_id}/withdraw`
    *   Input: `user_id`, `amount`, `currency`, `destination_details` (e.g., bank account info).
    *   Logic:
        *   Validate input, user/wallet existence, and sufficient wallet balance (calculated from ledger entries).
        *   Create a "withdrawal" `Entry` with two `EntryLine`s: Debit User Wallet Account, Credit System Asset Account (often a "Withdrawal Clearing Account").
        *   Persist the `Entry` within a database transaction. This locks in the balance change in the ledger.
        *   Initiate payout via a payment gateway or banking system, referencing the ledger `Entry` ID.
        *   Handle asynchronous payout status updates (success/failure) and potentially create a compensating entry if the payout fails after the initial ledger entry was posted.

*   **Transfer:**
    *   API Endpoint: `POST /wallets/{sender_user_id}/transfer`
    *   Input: `sender_user_id`, `recipient_user_id`, `amount`, `currency`.
    *   Logic:
        *   Validate input, existence of sender and recipient wallets, and sufficient sender balance.
        *   Create a "transfer" `Entry` with two `EntryLine`s: Debit Sender Wallet Account, Credit Recipient Wallet Account.
        *   Persist the `Entry` within a single database transaction.

*   **Apply Fee:**
    *   Internal Function/API: Triggered by other operations or scheduled tasks.
    *   Input: `user_id`, `amount`, `currency`, `fee_type`.
    *   Logic:
        *   Validate input and user/wallet existence.
        *   Ensure sufficient balance if the fee is deducted from the wallet.
        *   Create a "fee" `Entry` with two `EntryLine`s: Debit User Wallet Account, Credit System Revenue Account.
        *   Persist the `Entry` within a database transaction.

**Implementation Plan:**

*   Create functions for each operation that encapsulate the entry creation and persistence logic.
*   Implement validation rules for each operation.
*   Integrate with the database layer to perform operations within transactions.

## 5. Balance Calculation

Wallet balances are not stored directly on the `Account` entity (to avoid race conditions and ensure the ledger is the single source of truth). Instead, the balance is derived from the sum of debits and credits for an account from all relevant entries up to a specific point in time.

**Implementation Plan:**

*   Create a function `CalculateBalance(accountID string, upToTime time.Time) (float64, error)` that queries the database for all `EntryLine`s associated with the `accountID` where the parent `Entry`'s `Date` is less than or equal to `upToTime`.
*   Sum the `Credit` amounts and subtract the `Debit` amounts from the retrieved `EntryLine`s.
*   Consider performance optimization for balance calculation on large ledgers:
    *   Materialized views in the database.
    *   Caching of frequently accessed balances.
    *   Implementing a system to track cumulative balances after certain checkpoints.

## 6. Concurrency Handling using Database Transactions

As highlighted in the general ledger plan, database transactions are the fundamental mechanism for ensuring data consistency and integrity in a concurrent environment.

**Implementation Plan:**

*   **Always Use Transactions for Writes:** Every operation that creates or modifies `Entry` or `EntryLine` data (deposit, withdrawal, transfer, fee application) *must* be wrapped in a database transaction.
*   **Atomicity:** The transaction ensures that either all changes related to an `Entry` (the entry itself and all its lines) are successfully committed to the database, or none of them are. If any part fails, the transaction is rolled back, leaving the ledger in a consistent state.
*   **Isolation:** Use appropriate database isolation levels (e.g., Read Committed, Repeatable Read) to prevent issues like dirty reads or lost updates when multiple operations access the same account or entries concurrently. For financial transactions, `Repeatable Read` or higher is often preferred to prevent phantom reads during balance checks within a transaction.
*   **Preventing Overdrafts:** When processing withdrawals or transfers, the check for sufficient balance *must* be done within the same database transaction that creates the debit entry line. This prevents a race condition where two concurrent withdrawals could both check the balance, find it sufficient, and then both proceed to debit the account, resulting in an overdraft. The database transaction and isolation level will ensure that one transaction acquires a lock or detects that the balance has changed before committing.
*   **Error Handling:** Implement robust error handling for database transaction failures, including retries for transient errors (e.g., deadlock) and clear error reporting for permanent failures.

## 7. Audit Trail and Immutability

The ledger system inherently provides a complete and immutable audit trail of all financial activity.

**Implementation Plan:**

*   **Immutability:** Once an `Entry` is successfully committed to the database, it should **never** be modified or deleted.
*   **Corrections:** Any errors or necessary adjustments must be handled by creating new, offsetting `Entry` records. For example, to reverse a mistaken deposit, create a new entry that debits the user's wallet and credits the system asset account.
*   **Querying for Audit:** Provide capabilities to query the ledger by account, user, date range, transaction type, etc., to reconstruct the history of any wallet or system account.

## 8. Integration Points

A fintech wallet system needs to integrate with various external services.

**Implementation Plan:**

*   **Payment Gateways:** Integrate with payment processors for handling deposits (inbound payments) and withdrawals (outbound payouts). This will involve using their APIs and handling webhooks for asynchronous status updates.
*   **Fraud Detection Systems:** Integrate with fraud detection services to screen transactions in real-time or near-real-time before posting entries.
*   **Reporting and Analytics:** Provide interfaces (APIs or direct database access for reporting tools) to access ledger data for generating financial reports and performing business analytics.
*   **User Management:** Integrate with the platform's user management system to link wallet accounts to specific users and retrieve user information.
*   **External Banking Systems:** Direct integration with banks for clearing and settlement purposes if not using a payment gateway as an intermediary.

## 9. Error Handling, Reconciliation, and Monitoring

Robust error handling, reconciliation processes, and monitoring are essential for operational stability and financial accuracy.

**Implementation Plan:**

*   **Comprehensive Error Handling:** Implement detailed error logging and handling for all operations, especially those involving external integrations or database transactions.
*   **Idempotency:** Design APIs and internal logic to be idempotent where possible, especially for operations triggered by external systems (like webhooks from payment gateways) to handle duplicate requests gracefully.
*   **Reconciliation:** Implement automated reconciliation processes to compare the ledger's state with external systems (e.g., bank statements, payment gateway reports) to identify discrepancies. Develop procedures for investigating and correcting discrepancies through compensating entries.
*   **Monitoring and Alerting:** Implement comprehensive monitoring of system health, transaction volume, error rates, and key financial metrics. Set up alerts for anomalies or critical failures.

## 10. Security Considerations

Protecting sensitive financial data is paramount.

**Implementation Plan:**

*   **Authentication and Authorization:** Implement strong authentication for all access to the ledger system's APIs and internal interfaces. Use a granular authorization model to ensure that only authorized users or services can perform specific actions (e.g., only the deposit service can create deposit entries).
*   **Data Encryption:** Encrypt sensitive data at rest (in the database) and in transit (using TLS/SSL).
*   **Input Validation:** Strictly validate all input to prevent injection attacks and ensure data integrity.
*   **Auditing Access:** Log all access to the ledger system and changes made for security auditing purposes.
*   **Secure Credential Management:** Securely manage API keys and database credentials.
*   **Regular Security Audits:** Conduct regular security audits and penetration testing.

This detailed plan provides a solid foundation for building a robust fintech wallet ledger system based on double-entry principles. Each section can be further broken down into smaller tasks during the implementation phase.