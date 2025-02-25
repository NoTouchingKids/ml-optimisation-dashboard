package handler

import (
	"context"
	"sync"
	"time"

	"backend/internal/types"
)

// WSManager manages active WebSocket connections and message buffering
type WSManager struct {
	// Buffered messages per client
	messageBuffers map[string][]types.LogRecord
	bufferSize     int
	mu             sync.RWMutex

	// Channel for new messages
	messageChan chan types.LogRecord
}

func NewWSManager(bufferSize int) *WSManager {
	manager := &WSManager{
		messageBuffers: make(map[string][]types.LogRecord),
		bufferSize:     bufferSize,
		messageChan:    make(chan types.LogRecord, 1000),
	}

	go manager.processMessages()
	return manager
}

// AddMessage adds a new message to the buffer
func (m *WSManager) AddMessage(record types.LogRecord) {
	m.messageChan <- record
}

// GetMessages retrieves and clears messages for a client
func (m *WSManager) GetMessages(clientID string) []types.LogRecord {
	m.mu.Lock()
	defer m.mu.Unlock()

	messages := m.messageBuffers[clientID]
	m.messageBuffers[clientID] = nil
	return messages
}

// processMessages handles incoming messages and manages buffers
func (m *WSManager) processMessages() {
	for msg := range m.messageChan {
		m.mu.Lock()
		// Initialize buffer if needed
		if m.messageBuffers[msg.ClientID] == nil {
			m.messageBuffers[msg.ClientID] = make([]types.LogRecord, 0, m.bufferSize)
		}

		// Add message to buffer
		buffer := m.messageBuffers[msg.ClientID]
		if len(buffer) >= m.bufferSize {
			// Remove oldest message if buffer is full
			buffer = buffer[1:]
		}
		buffer = append(buffer, msg)
		m.messageBuffers[msg.ClientID] = buffer
		m.mu.Unlock()
	}
}

// StartCleanup starts a periodic cleanup of old messages
func (m *WSManager) StartCleanup(ctx context.Context, maxAge time.Duration) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.cleanup(maxAge)
			}
		}
	}()
}

// cleanup removes old messages
func (m *WSManager) cleanup(maxAge time.Duration) {
	threshold := time.Now().Add(-maxAge).UnixNano()

	m.mu.Lock()
	defer m.mu.Unlock()

	for clientID, messages := range m.messageBuffers {
		var newMessages []types.LogRecord
		for _, msg := range messages {
			if msg.Timestamp > threshold {
				newMessages = append(newMessages, msg)
			}
		}
		m.messageBuffers[clientID] = newMessages
	}
}
