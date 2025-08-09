# Task List: Implementing Core Data Structures for Fintech Wallet Ledger

This document outlines the tasks required to implement the core data structures for the fintech wallet ledger system in Go. These structures form the foundation upon which all other components of the ledger will be built.

## 1. Define and Implement the `Account` Struct

This task involves defining the Go struct for representing a financial account, specifically tailored for the fintech wallet context.

*   Define a `struct` named `Account`.
*   Include the following fields with appropriate Go types and JSON tags:
    *   `ID` (string): A unique identifier for the account (e.g., a UUID).
    *   `Name` (string): A human-readable name for the account (e.g., "User's Primary Wallet").
    *   `Type` (AccountType): The type of account (e.g., Asset, Liability).
    *   `UserID` (string): The identifier of the user who owns this wallet account. This is a foreign key to the user management system.
    *   `Currency` (string): The currency of the account (e.g., "USD", "EUR").
    *   `CreatedAt` (time.Time): Timestamp for when the account was created.
    *   `UpdatedAt` (time.Time): Timestamp for when the account was last updated.
    *   Consider adding a `Status` field (e.g., "active", "inactive", "closed").
*   Add necessary comments to explain the purpose of each field.

## 2. Define the `AccountType` Enum

This task involves defining the enumeration for different types of accounts.

*   Define a Go type `AccountType` as an alias for `string`.
*   Define constants for the standard account types (Asset, Liability, Equity, Revenue, Expense).
*   Consider adding fintech-specific account types if needed (though sticking to standard types is generally recommended for double-entry).
*   Add comments to explain the purpose of each account type.

## 3. Define and Implement the `Entry` Struct

This task involves defining the Go struct for representing a single financial transaction (an entry in the ledger).

*   Define a `struct` named `Entry`.
*   Include the following fields with appropriate Go types and JSON tags, considering its potential role within a CTE:
    *   `ID` (string): A unique identifier for the entry (e.g., a UUID).
    *   `Description` (string): A description of the transaction.
    *   `Date` (time.Time): The date and time the transaction occurred.
    *   `Lines` ([]EntryLine): A slice of `EntryLine` structs representing the debit and credit sides of the entry.
    *   `TransactionType` (string): A field to categorize the type of transaction (e.g., "deposit", "withdrawal", "transfer", "fee"). This helps in reporting and processing.
    *   `Source` (string): An identifier for the source of the transaction (e.g., the ID of the originating user, an external system identifier).
    *   `Status` (string): The current status of the entry (e.g., "pending", "posted", "failed", "reversed").
    *   `CreatedAt` (time.Time): Timestamp for when the entry was created.
    *   `UpdatedAt` (time.Time): Timestamp for when the entry was last updated.
    *   Consider adding a `CTEID` (string) field to link this entry to a specific Chained Transaction Event.
*   Add necessary comments to explain the purpose of each field.

## 4. Define and Implement the `EntryLine` Struct

This task involves defining the Go struct for representing a single line within an entry, linking to an account and specifying a debit or credit amount.

*   Define a `struct` named `EntryLine`.
*   Include the following fields with appropriate Go types and JSON tags:
    *   `AccountID` (string): The identifier of the account affected by this line. This is a foreign key to the `Account` struct.
    *   `Debit` (float64): The debit amount for this line. Should be non-negative.
    *   `Credit` (float64): The credit amount for this line. Should be non-negative.
    *   Ensure that either `Debit` or `Credit` is zero for any given line, and that debits and credits balance across all lines in an `Entry`.
*   Add necessary comments to explain the purpose of each field.

## 5. Define and Implement Data Structures for Chained Transaction Events (CTE)

This task involves defining the Go structs required to represent and manage Chained Transaction Events and their components.

*   Define a `struct` named `ChainedTransactionEvent`.
*   Include fields to track:
    *   `ID` (string): Unique identifier for the CTE.
    *   `State` (string): The current state of the CTE (e.g., "CREATED", "VALIDATED", "EXECUTING", "COMPLETED", "FAILED", "ROLLING_BACK", "ROLLED_BACK"). Define corresponding constants for these states.
    *   `Description` (string): A description of the overall chained event.
    *   `CreatedAt` (time.Time), `UpdatedAt` (time.Time).
    *   `Transactions` ([]CTETransaction): A slice of structs representing the individual transactions within the chain.
    *   Consider fields for tracking progress, errors, and completion time.
*   Define a `struct` named `CTETransaction`. This will represent a single step or transaction within a CTE.
*   Include fields to link to the actual ledger `Entry` (once created), define dependencies, and include compensation logic:
    *   `ID` (string): Unique identifier for this step within the CTE.
    *   `Description` (string): Description of this specific transaction step.
    *   `EntryID` (string): The ID of the `Entry` created in the core ledger for this transaction (optional, will be populated upon execution).
    *   `State` (string): State of this individual transaction step (e.g., "PENDING", "EXECUTING", "COMPLETED", "FAILED", "COMPENSATING", "COMPENSATED"). Define constants.
    *   `Dependencies` ([]string): A list of `CTETransaction` IDs that must complete before this one can start.
    *   `Compensation` (CompensationDetails): Details required to compensate this transaction if needed.
    *   Consider fields for conditional execution and parallel group membership if those features are implemented in this stage.
*   Define a `struct` named `CompensationDetails`.
*   Include fields to specify how to compensate for this transaction:
    *   `Type` (string): Type of compensation (e.g., "create_offsetting_entry", "call_external_service").
    *   `EntryTemplate` (*Entry): An optional template for creating an offsetting ledger entry for compensation.
    *   `ExternalCallDetails` (*ExternalCall): Details for calling an external service for compensation.

## 6. Define and Implement Data Structures for Chained Transaction Event Lien (CTEL)

This task involves defining the Go structs required to represent and manage CTELs.

*   Define a `struct` named `CTELEntry`.
*   Include fields to track:
    *   `ID` (string): Unique identifier for the CTEL.
    *   `CTEID` (string): The ID of the Chained Transaction Event this lien is associated with.
    *   `AccountID` (string): The account from which funds are being liened.
    *   `Amount` (float64): The amount of funds under lien.
    *   `State` (string): The state of the CTEL (e.g., "CREATED", "ACTIVE", "UTILIZED", "EXPIRED", "SETTLING", "RELEASED"). Define constants.
    *   `CreatedAt` (time.Time), `UpdatedAt` (time.Time), `ExpiresAt` (time.Time).

## 7. Review and Refine All Data Structures

*   Review all defined structs and the `AccountType` enum for clarity, completeness, and correctness.
*   Ensure consistent naming conventions and data types.
*   Verify that the relationships between structs (e.g., `Entry` containing `EntryLine`s, `EntryLine` referencing `AccountID`) are clear and support the double-entry principles.
*   Consider any additional fields that might be necessary for auditing, compliance, or specific fintech requirements.

This task list provides a clear roadmap for implementing the foundational data structures. Subsequent tasks will build upon these structures to implement the logic for creating, validating, and processing entries, managing accounts, and generating reports.