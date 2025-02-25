-- Create the logs table
CREATE TABLE
IF NOT EXISTS logs
(
    timestamp TIMESTAMPTZ NOT NULL,
    client_id TEXT NOT NULL,
    message BYTEA NOT NULL,
    process_id INTEGER,
    PRIMARY KEY
(timestamp, client_id)
);

-- Create the TimescaleDB hypertable
SELECT create_hypertable('logs', 'timestamp', if_not_exists
=> TRUE);

-- Create an index on client_id for faster queries
CREATE INDEX
IF NOT EXISTS idx_logs_client_id ON logs
(client_id);

-- Grant permissions
GRANT ALL PRIVILEGES ON TABLE logs TO postgres;