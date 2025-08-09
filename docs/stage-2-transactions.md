# Stage 2: Transaction and Entry Management

This document provides a detailed plan for implementing the robust transaction and entry management components of a production-ready double-entry financial ledger system. This stage is critical for capturing, validating, and persistently storing individual financial movements while ensuring data integrity.

## Data Structures

We will define the necessary data structures to represent transactions and the individual lines within them, adhering to the principles of double-entry bookkeeping.

### Entry

The `Entry` struct represents a single, atomic financial transaction. It encapsulates all the information related to a specific event that impacts the ledger.
It is the core unit of record in the ledger.

```
go
type Entry struct {
	ID          string    // Unique identifier for the entry
	Description string    // A brief description of the transaction
	Date        time.Time // The date the transaction occurred
	Lines       []EntryLine // A list of the individual entry lines that make up this transaction
	CreatedAt   time.Time // Timestamp for when the entry was recorded
	UpdatedAt   time.Time // Timestamp for when the entry was last modified
}
```
**Considerations:**

*   The `ID` will be a unique identifier, likely generated automatically (e.g., a UUID).
*   The `Date` field will be crucial for chronological ordering in journals and reports.
*   The `Lines` slice will contain at least two `EntryLine` elements to satisfy the double-entry requirement.
*   Timestamps (`CreatedAt`, `UpdatedAt`) will be included for auditing purposes.

### EntryLine

The `EntryLine` struct represents a single debit or credit entry affecting a specific account within an `Entry`.

```
go
type EntryLine struct {
	AccountID string  // The ID of the account affected by this line
	Debit     float64 // The debit amount (zero if it's a credit)
	Credit    float64 // The credit amount (zero if it's a debit)
}
```
**Considerations:**

*   Exactly one of `Debit` or `Credit` should be non-zero for a valid `EntryLine`.
*   The `AccountID` will link the entry line to the specific account defined in Stage 1.

## Core Functionality

This stage will involve implementing the functions necessary to create, manage, and validate entries.

### Creating New Entries

A function will be required to create a new `Entry`. This function will take the necessary details (description, date, and a list of entry line data) and construct the `Entry` object.

```
go
// CreateEntry creates a new financial entry.
func CreateEntry(description string, date time.Time, linesData []EntryLineData) (*Entry, error) {
	// ... implementation details
	return nil, nil // placeholder
}

// EntryLineData is a helper struct for creating entry lines.
type EntryLineData struct {
	AccountID string
	Debit     float64
	Credit    float64
}
```
**Implementation Plan:**

1.  Generate a unique `ID` for the new entry.
2.  Populate the `Description` and `Date` fields.
3.  Iterate through the provided `linesData` to create the `EntryLine` objects.
4.  Perform validation on the created entry.
5.  Set `CreatedAt` and `UpdatedAt` timestamps.
6.  Return the created `Entry` object or an error if validation fails.

### Validating Entries

A critical function will be responsible for validating an `Entry` to ensure it adheres to the principles of double-entry bookkeeping (total debits must equal total credits) and other business rules.
```
go
// ValidateEntry checks if an entry is valid (debits equal credits, etc.).
func ValidateEntry(entry *Entry) error {
	// ... implementation details
	return nil // placeholder
}
```
**Implementation Plan:**

1.  Iterate through the `EntryLine`s in the entry.
2.  Sum the total debit amounts.
3.  Sum the total credit amounts.
4.  Check if the total debits equal the total credits. If not, return an error indicating an imbalance.
5.  Add checks for other business rules (e.g., ensuring each `EntryLine` has a valid `AccountID`, that only one of debit/credit is non-zero per line).

### Storing and Retrieving Entries

While the storage mechanism (database) will be detailed in a later stage, functions for saving and retrieving entries will be designed here.
```
go
// SaveEntry persists an entry to the data store.
func SaveEntry(entry *Entry) error {
	// ... implementation details
	return nil // placeholder
}

// GetEntryByID retrieves an entry by its ID.
func GetEntryByID(entryID string) (*Entry, error) {
	// ... implementation details
	return nil, nil // placeholder
}

// GetEntriesByDateRange retrieves entries within a specified date range.
func GetEntriesByDateRange(startDate, endDate time.Time) ([]*Entry, error) {
	// ... implementation details
	return nil, nil // placeholder
}
```
**Implementation Plan:**

1.  Define the interfaces or functions for interacting with the data storage layer (e.g., `Save`, `GetByID`, `GetByDateRange`).
2.  Ensure data serialization/deserialization is handled correctly when interacting with the data store.
3.  Implement error handling for potential storage issues.

## Linking Entries to Accounts

The `AccountID` within the `EntryLine` struct provides the direct link between an entry and the accounts it affects. This link is fundamental for calculating account balances and generating reports in later stages.

**Implementation Plan:**

*   When creating `EntryLine`s, ensure the `AccountID` corresponds to an existing account in the system (validation may be performed here or during `ValidateEntry`).
*   When retrieving entries, the `EntryLine`s will implicitly link to the accounts through their `AccountID`.

## Storing Transaction Details

The `Entry.Description` field will store a brief summary of the transaction. For more detailed information, consider adding an optional field or a related structure if more complex transaction details are required (e.g., invoice numbers, customer information).

**Implementation Plan:**

*   Decide on the level of detail required in the `Description`.
*   If more detailed transaction data is needed, design a separate data structure and establish a relationship between `Entry` and this new structure (e.g., a `TransactionDetails` struct linked by the `Entry.ID`).