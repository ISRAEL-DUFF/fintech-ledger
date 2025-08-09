# Stage 3: Reporting

This document outlines the plan for implementing the reporting components of the double-entry financial ledger system. This stage focuses on generating standard financial reports based on the recorded account and transaction data.

## 3.1 Trial Balance

The trial balance is a list of all the debit and credit balances of accounts in the ledger at a specific point in time. It's used to verify that the total of all debit balances equals the total of all credit balances, a fundamental principle of double-entry bookkeeping.

### 3.1.1 Data Required

To generate a trial balance, we need access to:

*   All active accounts.
*   All entries up to the reporting date.

### 3.1.2 Logic for Generation

1.  **Iterate through Accounts:** Go through each account in the system.
2.  **Calculate Account Balance:** For each account, sum the debit amounts from all associated entry lines and subtract the sum of the credit amounts from all associated entry lines. Consider entries up to the reporting date.
3.  **Determine Balance Type:** If the calculated balance is positive, it's a debit balance. If it's negative, it's a credit balance (and the absolute value should be used).
4.  **Format Output:** Present the data as a table or list showing the account name, account type, debit balance, and credit balance for each account.
5.  **Verify Equality:** Calculate the total of all debit balances and the total of all credit balances. These totals should be equal.

### 3.1.3 Data Retrieval Strategy

To efficiently generate the trial balance, especially with a large number of entries, we should leverage database indexing on the `account_id` and `date` columns of the entries/entry_lines table. When querying for balances up to a specific date, filter by date first and then group by `account_id` to sum the debit and credit amounts. This aggregation should be performed within the database query for performance.

### 3.1.4 API Design

A dedicated API endpoint (e.g., `/reports/trial-balance`) should accept parameters for the reporting date. The API will trigger the report generation logic and return the trial balance data in a structured format (e.g., JSON).


## 3.2 Income Statement (Profit and Loss Statement)

The income statement reports a company's financial performance over a specific period. It summarizes revenues, expenses, and the resulting net income or net loss.

### 3.2.1 Data Required

To generate an income statement, we need access to:

*   Accounts categorized as Revenue and Expense types.
*   Entries within the reporting period.

### 3.2.2 Logic for Generation

1.  **Filter Accounts:** Select only accounts with the types Revenue and Expense.
2.  **Calculate Revenue Totals:** For each Revenue account, sum the credit amounts from all associated entry lines within the reporting period.
3.  **Calculate Expense Totals:** For each Expense account, sum the debit amounts from all associated entry lines within the reporting period.
4.  **Calculate Total Revenue:** Sum the totals from all Revenue accounts.
5.  **Calculate Total Expenses:** Sum the totals from all Expense accounts.
6.  **Calculate Net Income/Loss:** Subtract Total Expenses from Total Revenue. If the result is positive, it's Net Income. If negative, it's Net Loss.
7.  **Format Output:** Present the data showing a list of Revenue accounts and their totals, the total revenue, a list of Expense accounts and their totals, the total expenses, and the final Net Income or Net Loss.

## 3.3 Balance Sheet

### 3.2.3 Data Retrieval Strategy

Similar to the trial balance, efficient data retrieval for the income statement will rely on database indexing on `account_id` and `date`. Queries should filter entries within the specified reporting period and aggregate debit/credit sums grouped by `account_id`, specifically for accounts categorized as Revenue and Expense.

### 3.2.4 API Design

An API endpoint (e.g., `/reports/income-statement`) should accept parameters for the reporting period (start and end dates). It will return the income statement data in a structured format (e.g., JSON).


The balance sheet provides a snapshot of a company's financial position at a specific point in time. It follows the fundamental accounting equation: Assets = Liabilities + Equity.

### 3.3.1 Data Required

To generate a balance sheet, we need access to:

*   Accounts categorized as Asset, Liability, and Equity types.
*   Entries up to the reporting date.
*   The calculated Net Income or Loss from the income statement for the period ending at the balance sheet date (this is typically included in Equity).

### 3.3.2 Logic for Generation

1.  **Filter Accounts:** Select only accounts with the types Asset, Liability, and Equity.
2.  **Calculate Asset Totals:** For each Asset account, calculate the balance (debits - credits) up to the reporting date. Sum the balances of all Asset accounts to get Total Assets.
3.  **Calculate Liability Totals:** For each Liability account, calculate the balance (credits - debits) up to the reporting date. Sum the balances of all Liability accounts to get Total Liabilities.
4.  **Calculate Equity Totals (excluding Net Income):** For each Equity account (excluding any account specifically for Net Income), calculate the balance (credits - debits) up to the reporting date. Sum these balances.
5.  **Add Net Income/Loss to Equity:** Include the calculated Net Income (or subtract Net Loss) from the income statement to the total Equity.
6.  **Calculate Total Liabilities and Equity:** Sum Total Liabilities and the adjusted Total Equity.
7.  **Verify Accounting Equation:** Confirm that Total Assets equals Total Liabilities and Equity.
8.  **Format Output:** Present the data in sections for Assets, Liabilities, and Equity, showing individual account balances and the totals for each section.

## 3.4 Implementation Considerations

*   **Performance:** For large datasets, optimizing the retrieval and aggregation of entry data will be crucial.
*   **Reporting Periods:** The system needs to handle different reporting periods (monthly, quarterly, annually).
*   **Filtering and Sorting:** Users should be able to filter and sort report data.
*   **Exporting:** Reports should be exportable in common formats (e.g., CSV, PDF).
*   **Error Handling:** Implement robust error handling for cases where data inconsistencies might lead to imbalances in reports.