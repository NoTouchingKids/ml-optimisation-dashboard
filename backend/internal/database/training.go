package database

import (
	"context"
	"fmt"
)

func (c *Client) GetTrainingData(source string, page, limit int) ([]map[string]interface{}, int, error) {
	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", source)
	var total int
	err := c.pool.QueryRow(context.Background(), countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count query failed: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get paginated data
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1 OFFSET $2", source)
	rows, err := c.pool.Query(context.Background(), query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("data query failed: %w", err)
	}
	defer rows.Close()

	// Get column descriptions
	fieldDescriptions := rows.FieldDescriptions()
	numCols := len(fieldDescriptions)

	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, numCols)
		valuePtrs := make([]interface{}, numCols)
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, 0, fmt.Errorf("scanning row: %w", err)
		}

		row := make(map[string]interface{})
		for i, fd := range fieldDescriptions {
			row[string(fd.Name)] = values[i]
		}
		results = append(results, row)
	}

	return results, total, nil
}
