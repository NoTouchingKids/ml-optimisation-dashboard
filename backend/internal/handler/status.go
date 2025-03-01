package handler

import (
	"context"
	"log"
	"sync"

	"backend/internal/database"
	"backend/internal/event"
	"backend/internal/types"
)

// StatusHandler manages model status updates
type StatusHandler struct {
	db           *database.Client
	wsHandler    *WebSocketHandler
	consumer     *event.Consumer
	clientStatus map[string]types.ModelStatus
	mu           sync.RWMutex
}

// NewStatusHandler creates a new status handler
func NewStatusHandler(db *database.Client, wsHandler *WebSocketHandler, consumer *event.Consumer) *StatusHandler {
	handler := &StatusHandler{
		db:           db,
		wsHandler:    wsHandler,
		consumer:     consumer,
		clientStatus: make(map[string]types.ModelStatus),
	}

	// Subscribe to status events
	consumer.Subscribe(event.EventTypeModelStarted, handler.handleStatusUpdate)
	consumer.Subscribe(event.EventTypeModelCompleted, handler.handleStatusUpdate)
	consumer.Subscribe(event.EventTypeModelFailed, handler.handleStatusUpdate)
	consumer.Subscribe(event.EventTypeModelProgress, handler.handleStatusUpdate)

	return handler
}

// Start begins handling status updates
func (h *StatusHandler) Start(ctx context.Context) {
	h.consumer.Start(ctx)
}

// Stop halts status update handling
func (h *StatusHandler) Stop() {
	h.consumer.Stop()
}

// handleStatusUpdate processes a status update event
func (h *StatusHandler) handleStatusUpdate(ctx context.Context, eventType event.EventType, data []byte) error {
	// Deserialize event
	e, err := event.Deserialize(data, eventType)
	if err != nil {
		log.Printf("Error deserializing status event: %v", err)
		return err
	}

	statusEvent, ok := e.(*event.ModelStatusEvent)
	if !ok {
		log.Printf("Expected ModelStatusEvent but got %T", e)
		return nil
	}

	// Map event type to status
	status := statusEvent.Status
	if status == "" {
		switch eventType {
		case event.EventTypeModelStarted:
			status = "running"
		case event.EventTypeModelCompleted:
			status = "completed"
		case event.EventTypeModelFailed:
			status = "error"
		case event.EventTypeModelProgress:
			status = "running"
		}
	}

	// Update status in database
	modelStatus := types.ModelStatus{
		ClientID:    statusEvent.ClientID,
		Status:      status,
		Message:     statusEvent.Message,
		Timestamp:   statusEvent.Timestamp,
		ProcessType: statusEvent.ProcessType,
	}

	// Update database
	if err := h.db.UpdateModelStatus(ctx, modelStatus); err != nil {
		log.Printf("Failed to update model status in database: %v", err)
		// Continue anyway to update WebSocket clients
	}

	// Cache status
	h.mu.Lock()
	h.clientStatus[statusEvent.ClientID] = modelStatus
	h.mu.Unlock()

	// Broadcast to WebSocket clients
	h.broadcastStatus(modelStatus)

	return nil
}

// broadcastStatus sends a status update to relevant WebSocket clients
func (h *StatusHandler) broadcastStatus(status types.ModelStatus) {
	msg := types.WSMessage{
		Type:    types.MessageTypeModelStatus,
		Payload: status,
	}

	h.wsHandler.BroadcastToClient(status.ClientID, msg)
}

// GetStatus retrieves the current status for a client
func (h *StatusHandler) GetStatus(clientID string) (types.ModelStatus, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	status, exists := h.clientStatus[clientID]
	return status, exists
}
