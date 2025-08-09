# Stage 2: Account Management for Fintech Wallets

This document outlines the tasks required to implement the account management components specifically tailored for a fintech wallet ledger system. This stage builds upon the core data structures defined in Stage 1.

## Tasks

### 2.1 Define and Refine Account Data Structure

- [ ] **Task:** Review and finalize the `Account` Go struct based on the core data structures defined in Stage 1.
    - Ensure inclusion of essential fintech-specific fields: `ID`, `UserID`, `Currency`, `Type` (linking to `AccountType`), `Name`, `CreatedAt`, `UpdatedAt`.
    - Consider additional metadata fields as needed (e.g., status - active/inactive).
- [ ] **Task:** Finalize the `AccountType` enumeration, including types relevant to fintech wallets (e.g., `UserWallet`, `SystemHolding`, `SystemRevenue`, `SystemExpense`, `SystemClearing`).

### 2.2 Design Account Database Schema

- [ ] **Task:** Design the database schema for the `accounts` table.
    - Define columns corresponding to the fields in the `Account` struct.
    - Choose appropriate data types for each column (e.g., UUID for IDs, VARCHAR for strings, DECIMAL/NUMERIC for currency amounts if stored in the account record - though balances are usually calculated, not stored in the account record itself for double-entry).
    - Define primary keys and indexes (e.g., on `ID`, `UserID`, `Currency` for efficient lookups).
    - Consider foreign key constraints if linking directly to a `users` table (though a simple `UserID` field might suffice depending on the user management system design).
- [ ] **Task:** Write the database migration script(s) to create the `accounts` table.

### 2.3 Implement Account Data Access Layer (Repository)

- [ ] **Task:** Create a Go package or module for the account repository.
- [ ] **Task:** Implement a function `CreateAccount(account Account) (*Account, error)`:
    - Inserts a new account record into the database.
    - Should ensure uniqueness constraints (e.g., a user doesn't have multiple default wallets of the same currency unless explicitly designed).
    - Returns the created account with its generated ID (if applicable).
- [ ] **Task:** Implement a function `GetAccountByID(id string) (*Account, error)`:
    - Retrieves an account from the database by its ID.
    - Returns an error if the account is not found.
- [ ] **Task:** Implement a function `GetAccountsByUserID(userID string) ([]Account, error)`:
    - Retrieves all accounts associated with a specific user ID.
- [ ] **Task:** Implement a function `UpdateAccount(account Account) (*Account, error)`:
    - Updates an existing account record in the database.
    - **Crucially:** Define and enforce restrictions on which fields can be updated. Account balances are derived and *not* directly updated through this function. Fields like `UserID` or `Currency` might be immutable after creation depending on business rules. Updates are typically limited to non-financial metadata (e.g., account name).
- [ ] **Task:** Implement a function `DeleteAccount(id string) error` (Soft Delete):
    - Implement a soft delete mechanism for accounts (e.g., using an `is_deleted` flag or a `deleted_at` timestamp in the database). Accounts with financial history should generally not be hard-deleted.
    - Ensure that soft-deleted accounts are excluded from retrieval functions unless specifically requested.

### 2.4 Implement Account Service/Business Logic

- [ ] **Task:** Update the account service to be aware of the CTE/CTEL engine.
    - Modify functions that retrieve account information (e.g., `GetAccountByID`, `GetAccountsByUserID`) to potentially accept a CTE context identifier.
    - Implement logic to calculate and return a "virtual balance" when a CTE context is provided, taking into account CTEL credits and debits associated with that CTE. This will likely involve querying the CTEL data structures managed by the CTE/CTEL engine.
- [ ] **Task:** Ensure that any operations that *do* modify core account metadata (fields allowed to be updated) are either handled directly if they don't impact financial state, or are initiated through the CTE engine if they are part of a larger workflow. **Note:** Direct account balance updates are strictly prohibited; all balance changes must occur through posted ledger entries managed by the transaction service and potentially orchestrated by the CTE engine.


- [ ] **Task:** Create a Go package or module for the account service layer.
- [ ] **Task:** Implement functions in the service layer that utilize the repository for account operations.
- [ ] **Task:** Implement validation logic for account creation and updates:
    - Validate required fields.
    - Validate `AccountType`.
    - Validate currency format.
    - Ensure uniqueness where necessary.
- [ ] **Task:** Implement logic for handling different account types. This might involve specific validation or behavior based on the account type.
- [ ] **Task:** Implement logic for associating accounts with users, ensuring that when a user is created or onboarded, necessary default wallet accounts are created.

### 2.5 Implement Account API Endpoints

- [ ] **Task:** Design and implement REST or gRPC API endpoints for account management.
    - Consider adding endpoints or parameters to existing endpoints to retrieve account balances within the context of a specific CTE (returning the virtual balance).
    - Ensure that API calls attempting to modify financial aspects of accounts (which should not be allowed) are rejected or routed through the appropriate transaction/CTE creation endpoints.


    - Endpoints for creating a new account.
    - Endpoints for retrieving an account by ID.
    - Endpoints for retrieving accounts by user ID.
    - Endpoints for updating an account (with restrictions).
    - Endpoints for (soft) deleting an account.
- [ ] **Task:** Implement request and response structures for the API.
- [ ] **Task:** Implement authentication and authorization for account API endpoints, ensuring users can only access or modify their own accounts (unless they have administrator privileges).

### 2.6 Implement Account Testing

- [ ] **Task:** Write unit and integration tests specifically for retrieving virtual balances in the context of CTEs, ensuring the logic correctly incorporates CTEL data.


- [ ] **Task:** Write unit tests for the account data access layer (repository) to test CRUD operations against a mock or in-memory database.
- [ ] **Task:** Write unit tests for the account service layer to test business logic and validation.
- [ ] **Task:** Write integration tests to test the account API endpoints and their interaction with the database.

### 2.7 Documentation

- [ ] **Task:** Update the `docs/fintech_wallets/plan.md` file to mark Stage 2 tasks as completed or in progress.
- [ ] **Task:** Add detailed documentation for the Account API endpoints, including request/response formats, authentication requirements, and error codes.
- [ ] **Task:** Document the database schema for the `accounts` table.