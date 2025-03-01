package event

import (
	"encoding/json"
	"time"
)

// EventType defines the type of event
type EventType string

const (
	// Command events
	EventTypeTrainRequested   EventType = "train.requested"
	EventTypePredictRequested EventType = "predict.requested"

	// Status events
	EventTypeModelStarted   EventType = "model.started"
	EventTypeModelCompleted EventType = "model.completed"
	EventTypeModelFailed    EventType = "model.failed"
	EventTypeModelProgress  EventType = "model.progress"
)

// BaseEvent contains common fields for all events
type BaseEvent struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	ClientID  string    `json:"client_id"`
}

// TrainRequestedEvent represents a model training request
type TrainRequestedEvent struct {
	BaseEvent
	Data          []float64   `json:"data,omitempty"`
	StartDate     string      `json:"start_date,omitempty"`
	EndDate       string      `json:"end_date,omitempty"`
	Configuration interface{} `json:"config,omitempty"`
}

// PredictRequestedEvent represents a prediction request
type PredictRequestedEvent struct {
	BaseEvent
	Data          []float64   `json:"data,omitempty"`
	Configuration interface{} `json:"config,omitempty"`
}

// ModelStatusEvent represents a status update from a model
type ModelStatusEvent struct {
	BaseEvent
	Status      string `json:"status"`
	Message     string `json:"message,omitempty"`
	ProcessType string `json:"process_type"`
	Progress    int    `json:"progress,omitempty"`
}

// Serialize converts an event to JSON bytes
func Serialize(event interface{}) ([]byte, error) {
	return json.Marshal(event)
}

// Deserialize converts JSON bytes to an event
func Deserialize(data []byte, eventType EventType) (interface{}, error) {
	var event interface{}

	switch eventType {
	case EventTypeTrainRequested:
		event = &TrainRequestedEvent{}
	case EventTypePredictRequested:
		event = &PredictRequestedEvent{}
	case EventTypeModelStarted, EventTypeModelCompleted, EventTypeModelFailed, EventTypeModelProgress:
		event = &ModelStatusEvent{}
	default:
		// For unknown event types, deserialize to a map
		event = &map[string]interface{}{}
	}

	err := json.Unmarshal(data, event)
	return event, err
}
