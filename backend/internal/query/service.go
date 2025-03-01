package query

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"backend/internal/database"
	"backend/internal/event"
)

// ModelStats contains derived statistics for model runs
type ModelStats struct {
	AverageRuntime    float64 `json:"average_runtime"`
	TotalRuns         int     `json:"total_runs"`
	SuccessRate       float64 `json:"success_rate"`
	LastRunTime       string  `json:"last_run_time"`
	AverageLogCount   float64 `json:"average_log_count"`
	MostCommonWarning string  `json:"most_common_warning"`
}

// ModelState represents the current state of a model
type ModelState struct {
	ClientID     string      `json:"client_id"`
	Status       string      `json:"status"`
	ProcessType  string      `json:"process_type"`
	StartTime    time.Time   `json:"start_time"`
	EndTime      *time.Time  `json:"end_time,omitempty"`
	Runtime      float64     `json:"runtime"`
	Message      string      `json:"message"`
	Config       interface{} `json:"config,omitempty"`
	Performance  interface{} `json:"performance,omitempty"`
	Stats        ModelStats  `json:"stats,omitempty"`
	LogCount     int         `json:"log_count"`
	ErrorCount   int         `json:"error_count"`
	WarningCount int         `json:"warning_count"`
}

// QueryFilter provides filtering options for queries
type QueryFilter struct {
	ClientID      string    `json:"client_id"`
	ProcessType   string    `json:"process_type"`
	Status        string    `json:"status"`
	StartTimeFrom time.Time `json:"start_time_from"`
	StartTimeTo   time.Time `json:"start_time_to"`
	Limit         int       `json:"limit"`
	Offset        int       `json:"offset"`
}

// QueryService maintains materialized views and provides query APIs
type QueryService struct {
	db            *database.Client
	consumer      *event.Consumer
	runningModels map[string]*ModelState
	modelHistory  map[string][]*ModelState
	mu            sync.RWMutex
	running       bool
	stopChan      chan struct{}
}

// NewQueryService creates a new query service
func NewQueryService(db *database.Client, consumer *event.Consumer) *QueryService {
	return &QueryService{
		db:            db,
		consumer:      consumer,
		runningModels: make(map[string]*ModelState),
		modelHistory:  make(map[string][]*ModelState),
		stopChan:      make(chan struct{}),
	}
}

// Start begins the query service
func (s *QueryService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.mu.Unlock()

	// Subscribe to model status events
	s.consumer.Subscribe(event.EventTypeModelStarted, s.handleModelStatusUpdate)
	s.consumer.Subscribe(event.EventTypeModelCompleted, s.handleModelStatusUpdate)
	s.consumer.Subscribe(event.EventTypeModelFailed, s.handleModelStatusUpdate)
	s.consumer.Subscribe(event.EventTypeModelProgress, s.handleModelStatusUpdate)

	// Start the Kafka consumer
	s.consumer.Start(ctx)

	// Load initial state from database
	go s.loadInitialState(ctx)

	// Start periodic refresh of statistics
	go s.periodicRefresh(ctx)

	log.Println("Query service started")
	return nil
}

// Stop halts the query service
func (s *QueryService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopChan)
	s.running = false
}

// loadInitialState loads the current state from the database
func (s *QueryService) loadInitialState(ctx context.Context) {
	// Load model statuses from database
	statuses, err := s.db.GetAllModelStatuses(ctx)
	if err != nil {
		log.Printf("Error loading initial model states: %v", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, status := range statuses {
		modelState := &ModelState{
			ClientID:    status.ClientID,
			Status:      status.Status,
			ProcessType: status.ProcessType,
			StartTime:   status.Timestamp,
			Message:     status.Message,
		}

		// If status is running, add to running models
		if status.Status == "running" || status.Status == "pending" {
			s.runningModels[status.ClientID] = modelState
		}

		// Add to history for this client
		s.modelHistory[status.ClientID] = append(s.modelHistory[status.ClientID], modelState)
	}

	// For each client, calculate logs count
	for clientID, state := range s.runningModels {
		logCount, err := s.db.CountClientLogs(ctx, clientID, state.StartTime, time.Now())
		if err != nil {
			log.Printf("Error counting logs for client %s: %v", clientID, err)
			continue
		}
		state.LogCount = logCount
	}
}

// periodicRefresh updates statistics periodically
func (s *QueryService) periodicRefresh(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.refreshModelStats(ctx)
		}
	}
}

// refreshModelStats updates statistics for all models
func (s *QueryService) refreshModelStats(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update stats for all clients
	for clientID, history := range s.modelHistory {
		if len(history) == 0 {
			continue
		}

		var totalRuntime float64
		successCount := 0
		totalRuns := len(history)
		totalLogs := 0

		for _, state := range history {
			// Calculate runtime
			var runtime float64
			if state.EndTime != nil {
				runtime = state.EndTime.Sub(state.StartTime).Seconds()
			} else if state.Status == "completed" || state.Status == "error" {
				// If status is final but no end time, use current as approximation
				runtime = time.Now().Sub(state.StartTime).Seconds()
			}
			totalRuntime += runtime

			// Count successes
			if state.Status == "completed" {
				successCount++
			}

			// Add log count
			totalLogs += state.LogCount
		}

		// Get the most recent model state
		latestState := history[len(history)-1]

		// Update statistics
		latestState.Stats = ModelStats{
			AverageRuntime:  totalRuntime / float64(totalRuns),
			TotalRuns:       totalRuns,
			SuccessRate:     float64(successCount) / float64(totalRuns) * 100,
			LastRunTime:     latestState.StartTime.Format(time.RFC3339),
			AverageLogCount: float64(totalLogs) / float64(totalRuns),
		}

		// Update running model if this client has one
		if running, ok := s.runningModels[clientID]; ok {
			running.Stats = latestState.Stats
		}
	}
}

// handleModelStatusUpdate processes model status update events
func (s *QueryService) handleModelStatusUpdate(ctx context.Context, eventType event.EventType, data []byte) error {
	// Deserialize event
	e, err := event.Deserialize(data, eventType)
	if err != nil {
		return fmt.Errorf("deserializing status event: %w", err)
	}

	statusEvent, ok := e.(*event.ModelStatusEvent)
	if !ok {
		return fmt.Errorf("expected ModelStatusEvent but got %T", e)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Handle based on the event type
	switch eventType {
	case event.EventTypeModelStarted:
		modelState := &ModelState{
			ClientID:    statusEvent.ClientID,
			Status:      "running",
			ProcessType: statusEvent.ProcessType,
			StartTime:   statusEvent.Timestamp,
			Message:     statusEvent.Message,
		}
		s.runningModels[statusEvent.ClientID] = modelState
		s.modelHistory[statusEvent.ClientID] = append(s.modelHistory[statusEvent.ClientID], modelState)

	case event.EventTypeModelCompleted, event.EventTypeModelFailed:
		// Try to find the running model
		if existing, ok := s.runningModels[statusEvent.ClientID]; ok {
			// Update status
			if eventType == event.EventTypeModelCompleted {
				existing.Status = "completed"
			} else {
				existing.Status = "error"
			}
			existing.Message = statusEvent.Message

			// Set end time
			now := time.Now()
			existing.EndTime = &now
			existing.Runtime = now.Sub(existing.StartTime).Seconds()

			// Remove from running models
			delete(s.runningModels, statusEvent.ClientID)
		} else {
			// If not found in running, create a new state for history
			status := ""
			if eventType == event.EventTypeModelCompleted {
				status = "completed"
			} else {
				status = "error"
			}
			modelState := &ModelState{
				ClientID:    statusEvent.ClientID,
				Status:      status,
				ProcessType: statusEvent.ProcessType,
				StartTime:   statusEvent.Timestamp,
				Message:     statusEvent.Message,
				EndTime:     &statusEvent.Timestamp,
			}
			s.modelHistory[statusEvent.ClientID] = append(s.modelHistory[statusEvent.ClientID], modelState)
		}

	case event.EventTypeModelProgress:
		// Update progress on running model
		if existing, ok := s.runningModels[statusEvent.ClientID]; ok {
			existing.Message = statusEvent.Message
		}
	}

	return nil
}

// GetModelState returns the current state of a specific model
func (s *QueryService) GetModelState(clientID string) (*ModelState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// First check running models
	if state, ok := s.runningModels[clientID]; ok {
		return state, true
	}

	// Then check history for the latest state
	if history, ok := s.modelHistory[clientID]; ok && len(history) > 0 {
		return history[len(history)-1], true
	}

	return nil, false
}

// GetRunningModels returns all currently running models
func (s *QueryService) GetRunningModels() []*ModelState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*ModelState, 0, len(s.runningModels))
	for _, state := range s.runningModels {
		result = append(result, state)
	}
	return result
}

// QueryModelHistory returns filtered model history
func (s *QueryService) QueryModelHistory(filter QueryFilter) []*ModelState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*ModelState

	// If client ID is specified, only search that client
	if filter.ClientID != "" {
		history, ok := s.modelHistory[filter.ClientID]
		if !ok {
			return nil
		}

		for _, state := range history {
			if s.matchesFilter(state, filter) {
				results = append(results, state)
			}
		}
	} else {
		// Search all clients
		for _, history := range s.modelHistory {
			for _, state := range history {
				if s.matchesFilter(state, filter) {
					results = append(results, state)
				}
			}
		}
	}

	// Sort and paginate results
	// (simplified - in production you'd want proper sorting and pagination)
	if filter.Limit > 0 && len(results) > filter.Limit {
		end := filter.Offset + filter.Limit
		if end > len(results) {
			end = len(results)
		}
		if filter.Offset < end {
			return results[filter.Offset:end]
		}
		return nil
	}

	return results
}

// matchesFilter checks if a model state matches the given filter
func (s *QueryService) matchesFilter(state *ModelState, filter QueryFilter) bool {
	if filter.ProcessType != "" && state.ProcessType != filter.ProcessType {
		return false
	}
	if filter.Status != "" && state.Status != filter.Status {
		return false
	}
	if !filter.StartTimeFrom.IsZero() && state.StartTime.Before(filter.StartTimeFrom) {
		return false
	}
	if !filter.StartTimeTo.IsZero() && state.StartTime.After(filter.StartTimeTo) {
		return false
	}
	return true
}

// GetLogSummary returns summarized log information for a specific client
func (s *QueryService) GetLogSummary(ctx context.Context, clientID string, from, to time.Time) (map[string]interface{}, error) {
	// Query database for log counts by level
	logCounts, err := s.db.GetLogCountsByLevel(ctx, clientID, from, to)
	if err != nil {
		return nil, fmt.Errorf("querying log counts: %w", err)
	}

	// Get log rate over time
	logRates, err := s.db.GetLogRateOverTime(ctx, clientID, from, to, 10) // 10 time buckets
	if err != nil {
		return nil, fmt.Errorf("querying log rates: %w", err)
	}

	return map[string]interface{}{
		"log_counts": logCounts,
		"log_rates":  logRates,
		"from":       from,
		"to":         to,
		"total_logs": s.sumLogCounts(logCounts),
	}, nil
}

// sumLogCounts sums the log counts across all levels
func (s *QueryService) sumLogCounts(logCounts map[string]int) int {
	total := 0
	for _, count := range logCounts {
		total += count
	}
	return total
}
