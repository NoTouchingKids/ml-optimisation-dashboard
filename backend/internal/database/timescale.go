package database

import (
	"context"
	"fmt"

	"backend/internal/config"
	"backend/internal/types"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

func New(cfg config.DatabaseConfig) (*Client, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	pool, err := pgxpool.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return &Client{pool: pool}, nil
}

func (c *Client) FetchLogs(ctx context.Context, clientID string, from, to int64, limit int) ([]types.LogRecord, error) {
	query := `
        SELECT timestamp, client_id, message, process_id 
        FROM logs 
        WHERE client_id = $1 
        AND timestamp >= $2 
        AND timestamp <= $3 
        ORDER BY timestamp DESC
        LIMIT $4
    `

	rows, err := c.pool.Query(ctx, query, clientID, from, to, limit)
	if err != nil {
		return nil, fmt.Errorf("querying logs: %w", err)
	}
	defer rows.Close()

	var logs []types.LogRecord
	for rows.Next() {
		var log types.LogRecord
		if err := rows.Scan(&log.Timestamp, &log.ClientID, &log.Message, &log.ProcessID); err != nil {
			return nil, fmt.Errorf("scanning log row: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (c *Client) SaveLog(ctx context.Context, log types.LogRecord) error {
	query := `
        INSERT INTO logs (timestamp, client_id, message, process_id)
        VALUES ($1, $2, $3, $4)
    `

	_, err := c.pool.Exec(ctx, query, log.Timestamp, log.ClientID, log.Message, log.ProcessID)
	if err != nil {
		return fmt.Errorf("inserting log: %w", err)
	}

	return nil
}

func (c *Client) UpdateModelStatus(ctx context.Context, status types.ModelStatus) error {
	query := `
        INSERT INTO model_status (client_id, status, message, timestamp, process_type)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (client_id) DO UPDATE 
        SET status = $2, message = $3, timestamp = $4, process_type = $5
    `

	_, err := c.pool.Exec(ctx, query,
		status.ClientID,
		status.Status,
		status.Message,
		status.Timestamp,
		status.ProcessType,
	)
	if err != nil {
		return fmt.Errorf("updating model status: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	c.pool.Close()
	return nil
}

// Schema represents the database schema
func (c *Client) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS logs (
            timestamp   TIMESTAMPTZ NOT NULL,
            client_id  TEXT NOT NULL,
            message    BYTEA NOT NULL,
            process_id INTEGER,
            PRIMARY KEY (timestamp, client_id)
        )`,

		`SELECT create_hypertable('logs', 'timestamp', if_not_exists => TRUE)`,

		`CREATE TABLE IF NOT EXISTS model_status (
            client_id     TEXT PRIMARY KEY,
            status       TEXT NOT NULL,
            message      TEXT,
            timestamp    TIMESTAMPTZ NOT NULL,
            process_type TEXT NOT NULL
        )`,

		`CREATE INDEX IF NOT EXISTS idx_logs_client_id ON logs (client_id)`,
	}

	for _, query := range queries {
		if _, err := c.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("executing schema query: %w", err)
		}
	}

	return nil
}
