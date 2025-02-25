package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"backend/internal/buffer"
	"backend/internal/database"
	"backend/internal/grpc"
	"backend/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler manages WebSocket connections and message handling
type WebSocketHandler struct {
	// Connections management
	connections map[*types.WSConnection]bool
	mu          sync.RWMutex

	clients map[*types.WSConnection]bool

	// Dependencies
	db         *database.Client
	grpcClient *grpc.Client

	// WebSocket upgrader
	upgrader websocket.Upgrader

	logBuffer *buffer.LogBuffer
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(db *database.Client, grpcClient *grpc.Client, logBuffer *buffer.LogBuffer) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients:    make(map[*types.WSConnection]bool),
		logBuffer:  logBuffer,
		db:         db,
		grpcClient: grpcClient,
	}
}

// HandleConnection handles new WebSocket connections
func (h *WebSocketHandler) HandleConnection(c *gin.Context) {
	ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	clientID := c.Query("clientId")
	conn := &types.WSConnection{
		Conn:     ws,
		ClientID: clientID,
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	go h.handleMessages(conn)
	go h.streamLogs(conn)
}

func (h *WebSocketHandler) streamLogs(conn *types.WSConnection) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	noDataCount := 0
	maxEmptyChecks := 500 // Stop after ~5 seconds of no data

	for {
		select {
		case <-ticker.C:
			if !h.logBuffer.HasLogs(conn.ClientID) {

				noDataCount++
				if noDataCount >= maxEmptyChecks {
					return
				}
				continue
			}

			noDataCount = 0
			logs := h.logBuffer.PopLogs(conn.ClientID)

			if len(logs) > 0 {
				for _, logMsg := range logs {
					message := types.WSMessage{
						Type:    types.MessageTypeLiveLog,
						Payload: logMsg,
					}

					if err := conn.WriteJSON(message); err != nil {
						log.Printf("Error sending to websocket: %v", err)
						return
					}
				}
			}
		}
	}
}

// registerConnection adds a new WebSocket connection to the handler
func (h *WebSocketHandler) registerConnection(conn *types.WSConnection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[conn] = true

	// Setup cleanup when connection closes
	go func() {
		<-h.waitForDisconnect(conn)
		h.mu.Lock()
		delete(h.connections, conn)
		h.mu.Unlock()
		conn.Conn.Close()
	}()
}

// waitForDisconnect monitors connection for closure
func (h *WebSocketHandler) waitForDisconnect(conn *types.WSConnection) chan struct{} {
	done := make(chan struct{})
	go func() {
		for {
			if _, _, err := conn.Conn.NextReader(); err != nil {
				close(done)
				break
			}
		}
	}()
	return done
}

// handleMessages processes incoming WebSocket messages
func (h *WebSocketHandler) handleMessages(conn *types.WSConnection) {
	for {
		var message types.WSMessage
		err := conn.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) {
				log.Printf("WebSocket closed: %v", err)
			}
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
			return
		}

		switch message.Type {
		case types.MessageTypeHistoryReq:
			go h.handleHistoryRequest(conn, message)
		}
	}
}

// handleHistoryRequest processes requests for historical logs
func (h *WebSocketHandler) handleHistoryRequest(conn *types.WSConnection, msg types.WSMessage) {
	var req struct {
		FromTimestamp int64 `json:"from_timestamp"`
		ToTimestamp   int64 `json:"to_timestamp"`
		Limit         int   `json:"limit"`
	}

	if err := json.Unmarshal(msg.Payload.(json.RawMessage), &req); err != nil {
		h.sendErrorMessage(conn, "Invalid history request format")
		return
	}

	// Fetch logs from database
	logs, err := h.db.FetchLogs(context.Background(), conn.ClientID, req.FromTimestamp, req.ToTimestamp, req.Limit)
	if err != nil {
		h.sendErrorMessage(conn, "Failed to fetch history")
		return
	}

	// Send logs in batches
	batchSize := 100
	for i := 0; i < len(logs); i += batchSize {
		end := i + batchSize
		if end > len(logs) {
			end = len(logs)
		}

		response := types.WSMessage{
			Type:      types.MessageTypeLiveLog,
			RequestID: msg.RequestID,
			Payload:   logs[i:end],
		}

		if err := h.sendMessage(conn, response); err != nil {
			log.Printf("Failed to send history batch: %v", err)
			return
		}
	}
}

// handleModelStatus processes model status updates
func (h *WebSocketHandler) handleModelStatus(conn *types.WSConnection, msg types.WSMessage) {
	var status types.ModelStatus
	if err := json.Unmarshal(msg.Payload.(json.RawMessage), &status); err != nil {
		h.sendErrorMessage(conn, "Invalid model status format")
		return
	}

	// Update status in database
	if err := h.db.UpdateModelStatus(context.Background(), status); err != nil {
		log.Printf("Failed to update model status: %v", err)
		return
	}

	// Broadcast status to interested clients
	h.broadcastModelStatus(status)
}

// sendMessage sends a message to a specific connection
func (h *WebSocketHandler) sendMessage(conn *types.WSConnection, msg types.WSMessage) error {
	conn.Mu.Lock()
	defer conn.Mu.Unlock()
	return conn.Conn.WriteJSON(msg)
}

// sendErrorMessage sends an error message to a connection
func (h *WebSocketHandler) sendErrorMessage(conn *types.WSConnection, errMsg string) {
	response := types.WSMessage{
		Type: "error",
		Payload: map[string]string{
			"error": errMsg,
		},
	}
	h.sendMessage(conn, response)
}

// broadcastModelStatus sends status updates to all relevant connections
func (h *WebSocketHandler) broadcastModelStatus(status types.ModelStatus) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	msg := types.WSMessage{
		Type:    types.MessageTypeModelStatus,
		Payload: status,
	}

	for conn := range h.connections {
		if conn.ClientID == status.ClientID {
			go h.sendMessage(conn, msg)
		}
	}
}
