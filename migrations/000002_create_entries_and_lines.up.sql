-- +goose Up
-- SQL in this section is executed when the migration is applied

CREATE TABLE entries (
    id TEXT PRIMARY KEY,
    description TEXT NOT NULL,
    transaction_type TEXT NOT NULL,
    reference_id TEXT,
    status TEXT NOT NULL DEFAULT 'posted',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE entry_lines (
    id TEXT PRIMARY KEY,
    entry_id TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
    account_id TEXT NOT NULL REFERENCES accounts(id),
    debit DECIMAL(19,4) NOT NULL DEFAULT 0,
    credit DECIMAL(19,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL,
    FOREIGN KEY (account_id) REFERENCES accounts(id)
);

-- Add indexes for performance
CREATE INDEX idx_entry_lines_entry_id ON entry_lines(entry_id);
CREATE INDEX idx_entry_lines_account_id ON entry_lines(account_id);
CREATE INDEX idx_entries_created_at ON entries(created_at);
