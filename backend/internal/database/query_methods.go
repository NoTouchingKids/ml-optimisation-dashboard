package database

import (
	"context"
	"fmt"
	"time"

	"backend/internal/types"
)

// GetAllModelStatuses retrieves all model statuses
func (c *Client) GetAllModelStatuses(ctx context.Context) ([]types.ModelStatus, error) {
	query := `
		SELECT client_id, status, message, timestamp, process_type
		FROM model_status
		ORDER BY timestamp DESC
	`

	rows, err := c.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying model statuses: %w", err)
	}
	defer rows.Close()

	var statuses []types.ModelStatus
	for rows.Next() {
		var status types.ModelStatus
		if err := rows.Scan(
			&status.ClientID,
			&status.Status,
			&status.Message,
			&status.Timestamp,
			&status.ProcessType,
		); err != nil {
			return nil, fmt.Errorf("scanning model status: %w", err)
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// CountClientLogs counts logs for a client within a time range
func (c *Client) CountClientLogs(ctx context.Context, clientID string, from, to time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM logs
		WHERE client_id = $1
		AND timestamp >= $2
		AND timestamp <= $3
	`

	var count int
	err := c.pool.QueryRow(ctx, query, clientID, from, to).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting logs: %w", err)
	}

	return count, nil
}

// GetLogCountsByLevel returns log counts grouped by log level
func (c *Client) GetLogCountsByLevel(ctx context.Context, clientID string, from, to time.Time) (map[string]int, error) {
	// Note: This is a simplified implementation that would need to be adapted
	// to your actual log storage schema. Here, we're assuming you can extract
	// a log level from your binary log data.
	query := `
		WITH unpacked_logs AS (
			SELECT 
				client_id,
				timestamp,
				-- This would need to be adjusted based on your actual log format
				-- This is a placeholder for extracting log level from binary data
				'INFO' as level
			FROM logs
			WHERE client_id = $1
			AND timestamp >= $2
			AND timestamp <= $3
		)
		SELECT level, COUNT(*)
		FROM unpacked_logs
		GROUP BY level
	`

	rows, err := c.pool.Query(ctx, query, clientID, from, to)
	if err != nil {
		return nil, fmt.Errorf("querying log counts by level: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var level string
		var count int
		if err := rows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("scanning log level count: %w", err)
		}
		counts[level] = count
	}

	return counts, nil
}

// TimeSeriesBucket represents a time bucket with a count
type TimeSeriesBucket struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

// GetLogRateOverTime returns log counts in time buckets
func (c *Client) GetLogRateOverTime(ctx context.Context, clientID string, from, to time.Time, buckets int) ([]TimeSeriesBucket, error) {
	// Calculate bucket interval
	interval := to.Sub(from) / time.Duration(buckets)
	if interval < time.Second {
		interval = time.Second
	}

	// Create a time-bucket query using TimescaleDB's time_bucket function
	query := `
		SELECT 
			time_bucket($1, timestamp) AS bucket,
			COUNT(*) as count
		FROM logs
		WHERE client_id = $2
		AND timestamp >= $3
		AND timestamp <= $4
		GROUP BY bucket
		ORDER BY bucket
	`

	rows, err := c.pool.Query(ctx, query, interval.String(), clientID, from, to)
	if err != nil {
		return nil, fmt.Errorf("querying log rates: %w", err)
	}
	defer rows.Close()

	var timeSeries []TimeSeriesBucket
	for rows.Next() {
		var bucket time.Time
		var count int
		if err := rows.Scan(&bucket, &count); err != nil {
			return nil, fmt.Errorf("scanning time bucket: %w", err)
		}
		timeSeries = append(timeSeries, TimeSeriesBucket{
			Timestamp: bucket,
			Count:     count,
		})
	}

	return timeSeries, nil
}

// GetClientLogStats returns statistics about client logs
func (c *Client) GetClientLogStats(ctx context.Context, clientID string) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_logs,
			MIN(timestamp) as first_log,
			MAX(timestamp) as last_log,
			MAX(timestamp) - MIN(timestamp) as duration
		FROM logs
		WHERE client_id = $1
	`

	var totalLogs int
	var firstLog, lastLog time.Time
	var durationMicros int64

	err := c.pool.QueryRow(ctx, query, clientID).Scan(
		&totalLogs,
		&firstLog,
		&lastLog,
		&durationMicros,
	)
	if err != nil {
		return nil, fmt.Errorf("querying client log stats: %w", err)
	}

	// Convert microseconds to duration
	duration := time.Duration(durationMicros) * time.Microsecond

	// Calculate logs per second if duration is non-zero
	var logsPerSecond float64
	if duration > 0 {
		logsPerSecond = float64(totalLogs) / duration.Seconds()
	}

	return map[string]interface{}{
		"total_logs":     totalLogs,
		"first_log_time": firstLog,
		"last_log_time":  lastLog,
		"duration_sec":   duration.Seconds(),
		"logs_per_sec":   logsPerSecond,
	}, nil
}
