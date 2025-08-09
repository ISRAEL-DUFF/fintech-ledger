# Stage 5: Balance Calculation and Reporting

This document outlines the tasks required to implement balance calculation for wallet accounts and basic reporting functionalities for the fintech wallet ledger system.

## 5.1 Wallet Balance Calculation

### Task 5.1.1: Design and Implement Balance Calculation Function

*   **Description:** Create a Go function that calculates the balance of a given wallet account (`AccountID`) at a specific point in time (`endTime`).
*   **Implementation Details:**
    *   The function should query the database for all `EntryLine` records associated with the given `AccountID` with a date less than or equal to `endTime`.
    *   Iterate through the retrieved `EntryLine`s, summing up the `Credit` amounts and subtracting the `Debit` amounts.
    *   Handle potential performance issues for accounts with a very large number of entries (consider database indexing on AccountID and Date, or potentially caching or pre-calculated balances for frequently accessed accounts).
    *   Return the calculated balance as a `float64`.
*   **Dependencies:**
    *   Completed Stage 1 (Core Data Structures).
    *   Completed Stage 3 (Entry and Transaction Management, specifically the data access layer for entries).

### Task 5.1.2: Add API Endpoint for Wallet Balance

*   **Description:** Create an API endpoint that accepts a `UserID` and optionally an `endTime` and returns the balance of the user's primary wallet account (or a specified account).
*   **Implementation Details:**
    *   The endpoint should validate the input `UserID`.
    *   Retrieve the user's primary wallet account ID (or the specified AccountID).
    *   Call the balance calculation function (Task 5.1.1) with the appropriate AccountID and endTime.
    *   Return the balance in a suitable JSON format.
    *   Implement appropriate error handling (e.g., user not found, account not found).
*   **Dependencies:**
    *   Completed Task 5.1.1.
    *   API framework setup.

## 5.2 Basic Transaction Reporting

### Task 5.2.1: Design and Implement Transaction List Function

*   **Description:** Create a Go function that retrieves a list of transactions (`Entry` and associated `EntryLine`s) for a given wallet account (`AccountID`) within a specified date range (`startDate`, `endDate`).
*   **Implementation Details:**
    *   The function should query the database for `EntryLine` records associated with the given `AccountID` where the entry date is between `startDate` and `endDate` (inclusive).
    *   For each retrieved `EntryLine`, retrieve the corresponding `Entry` record and all its `EntryLine`s.
    *   Consider pagination for potentially large result sets.
    *   Order the results chronologically by entry date.
    *   Return a slice of `Entry` structs.
*   **Dependencies:**
    *   Completed Stage 1 (Core Data Structures).
    *   Completed Stage 3 (Entry and Transaction Management, specifically the data access layer for entries).

### Task 5.2.2: Add API Endpoint for Transaction List

*   **Description:** Create an API endpoint that accepts a `UserID`, `startDate`, and `endDate`, and returns a list of transactions for the user's primary wallet account within that date range.
*   **Implementation Details:**
    *   The endpoint should validate the input `UserID`, `startDate`, and `endDate`.
    *   Retrieve the user's primary wallet account ID.
    *   Call the transaction list function (Task 5.2.1) with the appropriate AccountID, startDate, and endDate.
    *   Return the list of transactions in a suitable JSON format.
    *   Implement appropriate error handling.
    *   Consider adding optional parameters for pagination and filtering.
*   **Dependencies:**
    *   Completed Task 5.2.1.
    *   API framework setup.

## 5.3 Testing

### Task 5.3.1: Write Unit Tests for Balance Calculation

*   **Description:** Write unit tests for the balance calculation function (Task 5.1.1) to verify its correctness with various scenarios, including:
    *   Accounts with only credits.
    *   Accounts with only debits.
    *   Accounts with both credits and debits.
    *   Calculating balance at different points in time.
    *   Empty accounts.
*   **Implementation Details:** Use a mocking framework or in-memory database for testing the function in isolation from the actual database.

### Task 5.3.2: Write Unit Tests for Transaction List

*   **Description:** Write unit tests for the transaction list function (Task 5.2.1) to verify its correctness with various scenarios, including:
    *   Retrieving transactions within a date range.
    *   Handling empty date ranges.
    *   Accounts with no transactions.
    *   Accounts with transactions outside the date range.
    *   Verifying the correct structure of returned entries and entry lines.
*   **Implementation Details:** Use a mocking framework or in-memory database for testing.

### Task 5.3.3: Write Integration Tests for Reporting API Endpoints

*   **Description:** Write integration tests to verify the functionality of the balance and transaction list API endpoints.
*   **Implementation Details:** These tests should interact with a test database, create test accounts and entries, and verify that the API endpoints return the expected data and handle errors correctly.

## 5.4 Documentation

### Task 5.4.1: Update API Documentation

*   **Description:** Document the new API endpoints for balance calculation and transaction listing in the API documentation (e.g., OpenAPI specification).
*   **Implementation Details:** Provide clear descriptions of the endpoints, required parameters, response formats, and potential error codes.

### Task 5.4.2: Update Internal Documentation

*   **Description:** Add detailed documentation for the balance calculation and transaction list functions, explaining their logic, parameters, and return values.
