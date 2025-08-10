-- Create the CTE events table
CREATE TABLE IF NOT EXISTS cte_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    state VARCHAR(20) NOT NULL DEFAULT 'CREATED',
    timeout INTERVAL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create an index on the state column for faster lookups
CREATE INDEX IF NOT EXISTS idx_cte_events_state ON cte_events (state);

-- Create the CTE transactions table
CREATE TABLE IF NOT EXISTS cte_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES cte_events(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(100) NOT NULL,
    state VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    "order" INTEGER NOT NULL,
    dependencies JSONB,
    payload JSONB,
    result JSONB,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_cte_transactions_event_id ON cte_transactions (event_id);
CREATE INDEX IF NOT EXISTS idx_cte_transactions_state ON cte_transactions (state);

-- Create the CTE liens table
CREATE TABLE IF NOT EXISTS cte_liens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES cte_events(id) ON DELETE CASCADE,
    account_id UUID NOT NULL,
    amount DECIMAL(20, 8) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    state VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_cte_liens_event_id ON cte_liens (event_id);
CREATE INDEX IF NOT EXISTS idx_cte_liens_account_id ON cte_liens (account_id);
CREATE INDEX IF NOT EXISTS idx_cte_liens_state ON cte_liens (state);
CREATE INDEX IF NOT EXISTS idx_cte_liens_expires_at ON cte_liens (expires_at);

-- Create a function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to automatically update the updated_at column
CREATE TRIGGER update_cte_events_updated_at
BEFORE UPDATE ON cte_events
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cte_transactions_updated_at
BEFORE UPDATE ON cte_transactions
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cte_liens_updated_at
BEFORE UPDATE ON cte_liens
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create a function to validate lien states
CREATE OR REPLACE FUNCTION validate_lien_state()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.state NOT IN ('PENDING', 'ACTIVE', 'RELEASED', 'EXPIRED') THEN
        RAISE EXCEPTION 'Invalid lien state: %', NEW.state;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to validate lien states
CREATE TRIGGER validate_lien_state_trigger
BEFORE INSERT OR UPDATE ON cte_liens
FOR EACH ROW
EXECUTE FUNCTION validate_lien_state();

-- Create a function to validate transaction states
CREATE OR REPLACE FUNCTION validate_transaction_state()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.state NOT IN ('PENDING', 'EXECUTING', 'COMPLETED', 'FAILED', 'COMPENSATED') THEN
        RAISE EXCEPTION 'Invalid transaction state: %', NEW.state;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to validate transaction states
CREATE TRIGGER validate_transaction_state_trigger
BEFORE INSERT OR UPDATE ON cte_transactions
FOR EACH ROW
EXECUTE FUNCTION validate_transaction_state();

-- Create a function to validate event states
CREATE OR REPLACE FUNCTION validate_event_state()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.state NOT IN ('CREATED', 'VALIDATING', 'VALIDATED', 'EXECUTING', 'COMPLETED', 'FAILED', 'ROLLING_BACK', 'ROLLED_BACK') THEN
        RAISE EXCEPTION 'Invalid event state: %', NEW.state;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create a trigger to validate event states
CREATE TRIGGER validate_event_state_trigger
BEFORE INSERT OR UPDATE ON cte_events
FOR EACH ROW
EXECUTE FUNCTION validate_event_state();
