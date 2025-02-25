package types

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message types
type WSMessageType string

const (
	MessageTypeLiveLog     WSMessageType = "live_log"
	MessageTypeModelStatus WSMessageType = "model_status"
	MessageTypeHistoryReq  WSMessageType = "history_request"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      WSMessageType `json:"type"`
	Payload   interface{}   `json:"payload,omitempty"`
	RequestID string        `json:"request_id,omitempty"`
}

// LogRecord represents a log entry
type LogRecord struct {
	Timestamp int64  `json:"timestamp"`
	ClientID  string `json:"client_id"`
	Message   []byte `json:"message"`
	ProcessID int32  `json:"process_id,omitempty"`
}

// WSConnection represents a WebSocket connection
type WSConnection struct {
	Conn     *websocket.Conn
	ClientID string
	Mu       sync.Mutex
}

func (c *WSConnection) WriteJSON(v interface{}) error {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	return c.Conn.WriteJSON(v)
}

// ModelStatus represents the current status of a model
type ModelStatus struct {
	ClientID    string    `json:"client_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	ProcessType string    `json:"process_type"`
}

// ModelRequest represents a request to train or run inference
type ModelRequest struct {
	Type          string      `json:"type"`
	ClientID      string      `json:"client_id"`
	Data          []float64   `json:"data,omitempty"`
	StartDate     string      `json:"start_date,omitempty"`
	EndDate       string      `json:"end_date,omitempty"`
	Configuration ModelConfig `json:"config,omitempty"`
}

// ModelConfig represents model configuration options
type ModelConfig struct {
	FeatureEngineering bool `json:"feature_engineering"`
	Detrend            bool `json:"detrend"`
	Difference         bool `json:"difference"`
}
