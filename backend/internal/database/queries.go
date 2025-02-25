package database

import (
	"context"
	"time"

	"backend/internal/types"
)

// LogQuery represents a query for logs
type LogQuery struct {
	ClientID string
	From     time.Time
	To       time.Time
	Limit    int
	Offset   int
}

func (c *Client) QueryLogs(ctx context.Context, q LogQuery) ([]types.LogRecord, error) {
	query := `
        SELECT timestamp, client_id, message, process_id 
        FROM logs 
        WHERE client_id = $1 
        AND timestamp >= $2 
        AND timestamp <= $3 
        ORDER BY timestamp DESC
        LIMIT $4 OFFSET $5
    `

	rows, err := c.pool.Query(ctx, query,
		q.ClientID, q.From, q.To, q.Limit, q.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []types.LogRecord
	for rows.Next() {
		var log types.LogRecord
		if err := rows.Scan(
			&log.Timestamp,
			&log.ClientID,
			&log.Message,
			&log.ProcessID,
		); err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (c *Client) GetModelStatus(ctx context.Context, clientID string) (*types.ModelStatus, error) {
	query := `
        SELECT status, message, timestamp, process_type
        FROM model_status
        WHERE client_id = $1
    `

	var status types.ModelStatus
	err := c.pool.QueryRow(ctx, query, clientID).Scan(
		&status.Status,
		&status.Message,
		&status.Timestamp,
		&status.ProcessType,
	)
	if err != nil {
		return nil, err
	}

	status.ClientID = clientID
	return &status, nil
}

func (c *Client) BatchInsertLogs(ctx context.Context, logs []types.LogRecord) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, log := range logs {
		_, err := tx.Exec(ctx, `
            INSERT INTO logs (timestamp, client_id, message, process_id)
            VALUES ($1, $2, $3, $4)
        `, log.Timestamp, log.ClientID, log.Message, log.ProcessID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
