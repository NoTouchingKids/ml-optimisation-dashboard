package buffer

import (
	"container/ring"
	"sync"

	"backend/internal/types"
)

type LogBuffer struct {
	mu      sync.RWMutex
	buffers map[string]*ring.Ring
	size    int
}

func NewLogBuffer(size int) *LogBuffer {
	return &LogBuffer{
		buffers: make(map[string]*ring.Ring),
		size:    size,
	}
}

func (lb *LogBuffer) Push(clientID string, record types.LogRecord) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	buffer, exists := lb.buffers[clientID]
	if !exists {
		buffer = ring.New(lb.size)
		lb.buffers[clientID] = buffer
	}

	buffer.Value = record
	lb.buffers[clientID] = buffer.Next()
}

func (lb *LogBuffer) GetLogs(clientID string) []types.LogRecord {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	buffer, exists := lb.buffers[clientID]
	if !exists || buffer == nil {
		return nil
	}

	var logs []types.LogRecord
	buffer.Do(func(x interface{}) {
		if x != nil {
			logs = append(logs, x.(types.LogRecord))
		}
	})

	return logs
}

func (lb *LogBuffer) PopLogs(clientID string) []types.LogRecord {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	buffer, exists := lb.buffers[clientID]
	if !exists || buffer == nil {
		return nil
	}

	var logs []types.LogRecord
	buffer.Do(func(x interface{}) {
		if x != nil {
			logs = append(logs, x.(types.LogRecord))
		}
	})

	// Clear the buffer
	lb.buffers[clientID] = ring.New(lb.size)

	return logs
}

func (lb *LogBuffer) HasLogs(clientID string) bool {
	lb.mu.RLock()
	defer lb.mu.RUnlock()

	buffer, exists := lb.buffers[clientID]
	if !exists || buffer == nil {
		return false
	}

	hasValue := false
	buffer.Do(func(x interface{}) {
		if x != nil {
			hasValue = true
		}
	})

	return hasValue
}
