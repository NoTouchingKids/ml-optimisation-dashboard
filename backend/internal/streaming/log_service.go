package streaming

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"backend/internal/buffer"
	"backend/internal/database"
	"backend/internal/types"
)

// LogStreamingService handles high-throughput log streaming
type LogStreamingService struct {
	// UDP server for receiving logs
	udpAddr     *net.UDPAddr
	udpConn     *net.UDPConn
	buffer      *buffer.LogBuffer
	db          *database.Client
	running     bool
	runningMu   sync.Mutex
	webSockets  WebSocketRegistry
	batchSize   int
	flushPeriod time.Duration
	stopChan    chan struct{}

	// Metrics tracking
	metrics   LogMetrics
	metricsMu sync.RWMutex
}

// LogMetrics tracks performance metrics for the log streaming service
type LogMetrics struct {
	LogsReceived   uint64
	LogsProcessed  uint64
	BytesReceived  uint64
	LastMinuteRate float64
	Errors         uint64
}

// WebSocketRegistry is an interface for registering WebSocket connections
type WebSocketRegistry interface {
	BroadcastToClient(clientID string, message types.WSMessage)
	GetClientConnections(clientID string) []*types.WSConnection
}

// NewLogStreamingService creates a new log streaming service
func NewLogStreamingService(
	udpHost string,
	udpPort int,
	bufferSize int,
	db *database.Client,
	webSockets WebSocketRegistry,
) (*LogStreamingService, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", udpHost, udpPort))
	if err != nil {
		return nil, fmt.Errorf("resolving UDP address: %w", err)
	}

	return &LogStreamingService{
		udpAddr:     addr,
		buffer:      buffer.NewLogBuffer(bufferSize),
		db:          db,
		webSockets:  webSockets,
		batchSize:   100,
		flushPeriod: 500 * time.Millisecond,
		stopChan:    make(chan struct{}),
		metrics:     LogMetrics{},
	}, nil
}

// Start begins the log streaming service
func (s *LogStreamingService) Start(ctx context.Context) error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if s.running {
		return nil
	}

	conn, err := net.ListenUDP("udp", s.udpAddr)
	if err != nil {
		return fmt.Errorf("starting UDP server: %w", err)
	}

	// Set buffer sizes for high throughput
	err = conn.SetReadBuffer(8 * 1024 * 1024) // 8MB read buffer
	if err != nil {
		log.Printf("Warning: failed to set UDP read buffer: %v", err)
	}

	s.udpConn = conn
	s.running = true

	// Start UDP receiver goroutine
	go s.receiveUDP(ctx)

	// Start periodic database persister
	go s.periodicallyPersistLogs(ctx)

	// Start metrics collector
	go s.collectMetrics(ctx)

	log.Printf("Log streaming service started on %s", s.udpAddr.String())
	return nil
}

// Stop halts the log streaming service
func (s *LogStreamingService) Stop() error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if !s.running {
		return nil
	}

	close(s.stopChan)
	s.running = false

	if s.udpConn != nil {
		return s.udpConn.Close()
	}

	return nil
}

// receiveUDP continuously receives UDP packets
func (s *LogStreamingService) receiveUDP(ctx context.Context) {
	buffer := make([]byte, 64*1024) // 64KB buffer for each packet

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		default:
			// Set read deadline to allow for checking stop conditions
			s.udpConn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

			n, addr, err := s.udpConn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// This is just a timeout, continue
					continue
				}

				s.incrementErrorCount()
				log.Printf("Error reading UDP: %v", err)
				continue
			}

			// Process the log message
			s.updateReceiveMetrics(n)
			go s.processLogMessage(buffer[:n], addr)
		}
	}
}

// processLogMessage handles a single log message
func (s *LogStreamingService) processLogMessage(data []byte, addr *net.UDPAddr) {
	// The first 36 bytes are the client ID
	if len(data) <= 36 {
		s.incrementErrorCount()
		log.Printf("Received malformed log message from %s: too short", addr.String())
		return
	}

	clientID := string(data[:36])
	message := data[36:]

	timestamp := time.Now().UnixNano()

	// Create a log record
	record := types.LogRecord{
		Timestamp: timestamp,
		ClientID:  clientID,
		Message:   message,
	}

	// Add to in-memory buffer
	s.buffer.Push(clientID, record)

	// Stream to WebSocket clients immediately
	s.streamToWebSocket(record)

	s.updateProcessedMetrics(1)
}

// streamToWebSocket sends a log record to connected WebSocket clients
func (s *LogStreamingService) streamToWebSocket(record types.LogRecord) {
	message := types.WSMessage{
		Type:    types.MessageTypeLiveLog,
		Payload: record,
	}

	// Broadcast to all connections for this client
	s.webSockets.BroadcastToClient(record.ClientID, message)
}

// periodicallyPersistLogs saves logs to the database at regular intervals
func (s *LogStreamingService) periodicallyPersistLogs(ctx context.Context) {
	ticker := time.NewTicker(s.flushPeriod)
	defer ticker.Stop()

	clientBatches := make(map[string][]types.LogRecord)

	for {
		select {
		case <-ctx.Done():
			s.flushAllLogs(clientBatches)
			return
		case <-s.stopChan:
			s.flushAllLogs(clientBatches)
			return
		case <-ticker.C:
			// Check all clients for logs to flush
			s.flushReadyBatches(clientBatches)
		}
	}
}

// flushReadyBatches persists batches of logs for each client
func (s *LogStreamingService) flushReadyBatches(clientBatches map[string][]types.LogRecord) {
	// Create a list of client IDs to check
	var clientIDs []string

	// First check existing batches
	for clientID := range clientBatches {
		clientIDs = append(clientIDs, clientID)
	}

	// Then check for new clients with logs
	// This could be optimized with a more direct buffer API

	// For each client, get and process logs
	for _, clientID := range clientIDs {
		// Add any new logs from buffer
		if s.buffer.HasLogs(clientID) {
			newLogs := s.buffer.GetLogs(clientID)
			clientBatches[clientID] = append(clientBatches[clientID], newLogs...)
		}

		// If we have enough logs, persist them
		if len(clientBatches[clientID]) >= s.batchSize {
			s.persistLogs(clientBatches[clientID])
			delete(clientBatches, clientID) // Clear after persistence
		}
	}
}

// flushAllLogs persists all remaining logs
func (s *LogStreamingService) flushAllLogs(clientBatches map[string][]types.LogRecord) {
	for _, logs := range clientBatches {
		if len(logs) > 0 {
			s.persistLogs(logs)
		}
	}
	// Clear all batches
	for clientID := range clientBatches {
		delete(clientBatches, clientID)
	}
}

// persistLogs saves a batch of logs to the database
func (s *LogStreamingService) persistLogs(logs []types.LogRecord) {
	if len(logs) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.db.BatchInsertLogs(ctx, logs); err != nil {
		s.incrementErrorCount()
		log.Printf("Error persisting logs: %v", err)
	}
}

// updateReceiveMetrics tracks the number of logs and bytes received
func (s *LogStreamingService) updateReceiveMetrics(bytes int) {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()

	s.metrics.LogsReceived++
	s.metrics.BytesReceived += uint64(bytes)
}

// updateProcessedMetrics tracks the number of logs processed
func (s *LogStreamingService) updateProcessedMetrics(count int) {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()

	s.metrics.LogsProcessed += uint64(count)
}

// incrementErrorCount tracks the number of errors
func (s *LogStreamingService) incrementErrorCount() {
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()

	s.metrics.Errors++
}

// collectMetrics periodically calculates performance metrics
func (s *LogStreamingService) collectMetrics(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	var lastLogsProcessed uint64
	var lastTimestamp time.Time = time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			now := time.Now()

			s.metricsMu.Lock()
			currentProcessed := s.metrics.LogsProcessed
			duration := now.Sub(lastTimestamp).Seconds()

			// Calculate logs per second over the last period
			if duration > 0 {
				logsInPeriod := currentProcessed - lastLogsProcessed
				s.metrics.LastMinuteRate = float64(logsInPeriod) / duration
			}

			lastLogsProcessed = currentProcessed
			lastTimestamp = now
			s.metricsMu.Unlock()

			// Log metrics periodically
			log.Printf("Log streaming metrics - Received: %d, Processed: %d, Rate: %.2f logs/sec, Errors: %d",
				s.GetMetrics().LogsReceived,
				s.GetMetrics().LogsProcessed,
				s.GetMetrics().LastMinuteRate,
				s.GetMetrics().Errors)
		}
	}
}

// GetMetrics returns the current metrics
func (s *LogStreamingService) GetMetrics() LogMetrics {
	s.metricsMu.RLock()
	defer s.metricsMu.RUnlock()

	// Return a copy to avoid race conditions
	return LogMetrics{
		LogsReceived:   s.metrics.LogsReceived,
		LogsProcessed:  s.metrics.LogsProcessed,
		BytesReceived:  s.metrics.BytesReceived,
		LastMinuteRate: s.metrics.LastMinuteRate,
		Errors:         s.metrics.Errors,
	}
}
