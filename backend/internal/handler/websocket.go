// backend/internal/handler/websocket.go
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

	clients   map[string][]*types.WSConnection
	clientsMu sync.RWMutex

	// Dependencies
	db         *database.Client
	grpcClient *grpc.Client

	// WebSocket upgrader
	upgrader websocket.Upgrader

	logBuffer *buffer.LogBuffer
}

// GetClientConnections returns all WebSocket connections for a client
func (h *WebSocketHandler) GetClientConnections(clientID string) []*types.WSConnection {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	// Return a copy of the connections to avoid race conditions
	connections := make([]*types.WSConnection, len(h.clients[clientID]))
	copy(connections, h.clients[clientID])
	return connections
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(db *database.Client, grpcClient *grpc.Client, logBuffer *buffer.LogBuffer) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		connections: make(map[*types.WSConnection]bool),
		clients:     make(map[string][]*types.WSConnection),
		logBuffer:   logBuffer,
		db:          db,
		grpcClient:  grpcClient,
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

	h.registerConnection(conn)
	go h.handleMessages(conn)
	go h.streamLogs(conn)
}

// BroadcastToClient sends a message to all connections for a specific client
func (h *WebSocketHandler) BroadcastToClient(clientID string, message types.WSMessage) {
	h.clientsMu.RLock()
	connections := h.clients[clientID]
	h.clientsMu.RUnlock()

	for _, conn := range connections {
		// Send in a non-blocking way
		go func(c *types.WSConnection, msg types.WSMessage) {
			if err := h.sendMessage(c, msg); err != nil {
				log.Printf("Error broadcasting to client %s: %v", clientID, err)
			}
		}(conn, message)
	}
}

// BroadcastToAll sends a message to all connected clients
func (h *WebSocketHandler) BroadcastToAll(message types.WSMessage) {
	h.mu.RLock()
	connections := make([]*types.WSConnection, 0, len(h.connections))
	for conn := range h.connections {
		connections = append(connections, conn)
	}
	h.mu.RUnlock()

	for _, conn := range connections {
		// Send in a non-blocking way
		go func(c *types.WSConnection, msg types.WSMessage) {
			if err := h.sendMessage(c, msg); err != nil {
				log.Printf("Error broadcasting to all: %v", err)
			}
		}(conn, message)
	}
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
	h.connections[conn] = true
	h.mu.Unlock()

	// Register with client ID
	h.clientsMu.Lock()
	h.clients[conn.ClientID] = append(h.clients[conn.ClientID], conn)
	h.clientsMu.Unlock()

	// Setup cleanup when connection closes
	go func() {
		<-h.waitForDisconnect(conn)
		h.unregisterConnection(conn)
	}()
}

// unregisterConnection removes a WebSocket connection
func (h *WebSocketHandler) unregisterConnection(conn *types.WSConnection) {
	h.mu.Lock()
	delete(h.connections, conn)
	h.mu.Unlock()

	h.clientsMu.Lock()
	clientConns := h.clients[conn.ClientID]
	for i, c := range clientConns {
		if c == conn {
			// Remove this connection
			h.clients[conn.ClientID] = append(clientConns[:i], clientConns[i+1:]...)
			break
		}
	}
	// If no more connections for this client, remove the client entry
	if len(h.clients[conn.ClientID]) == 0 {
		delete(h.clients, conn.ClientID)
	}
	h.clientsMu.Unlock()

	conn.Conn.Close()
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
			h.unregisterConnection(conn)
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
