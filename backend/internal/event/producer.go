package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// Producer is responsible for publishing events to Kafka
type Producer struct {
	commandWriter *kafka.Writer
	statusWriter  *kafka.Writer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string, commandTopic, statusTopic string) *Producer {
	commandWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        commandTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}

	statusWriter := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        statusTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}

	return &Producer{
		commandWriter: commandWriter,
		statusWriter:  statusWriter,
	}
}

// PublishTrainRequest publishes a train request event
func (p *Producer) PublishTrainRequest(ctx context.Context, clientID string, data []float64, startDate, endDate string, config interface{}) error {
	event := TrainRequestedEvent{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Type:      EventTypeTrainRequested,
			Timestamp: time.Now(),
			ClientID:  clientID,
		},
		Data:          data,
		StartDate:     startDate,
		EndDate:       endDate,
		Configuration: config,
	}

	return p.publishEvent(ctx, p.commandWriter, event)
}

// PublishPredictRequest publishes a predict request event
func (p *Producer) PublishPredictRequest(ctx context.Context, clientID string, data []float64, config interface{}) error {
	event := PredictRequestedEvent{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Type:      EventTypePredictRequested,
			Timestamp: time.Now(),
			ClientID:  clientID,
		},
		Data:          data,
		Configuration: config,
	}

	return p.publishEvent(ctx, p.commandWriter, event)
}

// PublishModelStatus publishes a model status event
func (p *Producer) PublishModelStatus(ctx context.Context, eventType EventType, clientID, status, message, processType string, progress int) error {
	event := ModelStatusEvent{
		BaseEvent: BaseEvent{
			ID:        uuid.New().String(),
			Type:      eventType,
			Timestamp: time.Now(),
			ClientID:  clientID,
		},
		Status:      status,
		Message:     message,
		ProcessType: processType,
		Progress:    progress,
	}

	return p.publishEvent(ctx, p.statusWriter, event)
}

// publishEvent serializes and publishes an event to Kafka
func (p *Producer) publishEvent(ctx context.Context, writer *kafka.Writer, event interface{}) error {
	data, err := Serialize(event)
	if err != nil {
		return fmt.Errorf("serializing event: %w", err)
	}

	// Extract event type and ID for message key
	var eventType, eventID string
	switch e := event.(type) {
	case TrainRequestedEvent:
		eventType = string(e.Type)
		eventID = e.ID
	case PredictRequestedEvent:
		eventType = string(e.Type)
		eventID = e.ID
	case ModelStatusEvent:
		eventType = string(e.Type)
		eventID = e.ID
	default:
		eventType = "unknown"
		eventID = uuid.New().String()
	}

	// Use client ID as the message key for partitioning
	key := []byte(eventID)

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: data,
		Headers: []kafka.Header{
			{Key: "event_type", Value: []byte(eventType)},
		},
	})

	if err != nil {
		return fmt.Errorf("publishing event: %w", err)
	}

	return nil
}

// Close closes all Kafka writers
func (p *Producer) Close() error {
	if err := p.commandWriter.Close(); err != nil {
		return err
	}
	return p.statusWriter.Close()
}
