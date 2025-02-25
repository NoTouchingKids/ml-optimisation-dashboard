-- Create logs table with TimescaleDB
CREATE TABLE
IF NOT EXISTS logs
(
    timestamp   TIMESTAMPTZ NOT NULL,
    client_id   TEXT NOT NULL,
    message     BYTEA NOT NULL,
    process_id  INTEGER,
    PRIMARY KEY
(timestamp, client_id)
);

-- Convert to hypertable
SELECT create_hypertable('logs', 'timestamp', if_not_exists
=> TRUE);

-- Create model status table
CREATE TABLE
IF NOT EXISTS model_status
(
    client_id     TEXT PRIMARY KEY,
    status        TEXT NOT NULL,
    message       TEXT,
    timestamp     TIMESTAMPTZ NOT NULL,
    process_type  TEXT NOT NULL
);

-- Create indexes
CREATE INDEX
IF NOT EXISTS idx_logs_client_id ON logs
(client_id);

-- Add compression policy (optional)
SELECT add_compression_policy('logs', INTERVAL '7 days'
);

-- Add retention policy (optional)
SELECT add_retention_policy('logs', INTERVAL '30 days'
);