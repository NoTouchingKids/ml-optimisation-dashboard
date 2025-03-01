package event

import (
	"context"
	"log"
	"strings"
	"sync"

	"github.com/segmentio/kafka-go"
)

// EventHandler is a function that processes an event
type EventHandler func(ctx context.Context, eventType EventType, data []byte) error

// Consumer subscribes to and processes Kafka events
type Consumer struct {
	readers       []*kafka.Reader
	running       bool
	runningMu     sync.Mutex
	eventHandlers map[EventType][]EventHandler
	handlersLock  sync.RWMutex
	stopChan      chan struct{}
	allowedTopics map[string]bool
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, groupID string, topics []string) *Consumer {
	allowedTopics := make(map[string]bool)
	for _, topic := range topics {
		allowedTopics[topic] = true
	}

	consumer := &Consumer{
		readers:       make([]*kafka.Reader, 0, len(topics)),
		eventHandlers: make(map[EventType][]EventHandler),
		stopChan:      make(chan struct{}),
		allowedTopics: allowedTopics,
	}

	for _, topic := range topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			GroupID:     groupID,
			Topic:       topic,
			MinBytes:    10e3, // 10KB
			MaxBytes:    10e6, // 10MB
			StartOffset: kafka.LastOffset,
		})
		consumer.readers = append(consumer.readers, reader)
	}

	return consumer
}

// Subscribe registers a handler for a specific event type
func (c *Consumer) Subscribe(eventType EventType, handler EventHandler) {
	c.handlersLock.Lock()
	defer c.handlersLock.Unlock()

	c.eventHandlers[eventType] = append(c.eventHandlers[eventType], handler)
}

// UnsubscribeAll removes all handlers for a specific event type
func (c *Consumer) UnsubscribeAll(eventType EventType) {
	c.handlersLock.Lock()
	defer c.handlersLock.Unlock()

	delete(c.eventHandlers, eventType)
}

// Start begins consuming messages from Kafka
func (c *Consumer) Start(ctx context.Context) {
	c.runningMu.Lock()
	if c.running {
		c.runningMu.Unlock()
		return
	}
	c.running = true
	c.runningMu.Unlock()

	// Start a goroutine for each reader
	var wg sync.WaitGroup
	for _, reader := range c.readers {
		wg.Add(1)
		go func(r *kafka.Reader) {
			defer wg.Done()
			c.consumeMessages(ctx, r)
		}(reader)
	}

	// Wait for all reader goroutines to finish
	go func() {
		wg.Wait()
		c.runningMu.Lock()
		c.running = false
		c.runningMu.Unlock()
	}()
}

// Stop halts all Kafka consumers
func (c *Consumer) Stop() {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()

	if !c.running {
		return
	}

	close(c.stopChan)
	c.running = false

	for _, reader := range c.readers {
		reader.Close()
	}
}

// consumeMessages processes messages from a specific Kafka reader
func (c *Consumer) consumeMessages(ctx context.Context, reader *kafka.Reader) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				// If context is canceled or reader is closed, exit gracefully
				if strings.Contains(err.Error(), "context canceled") ||
					strings.Contains(err.Error(), "reader closed") {
					return
				}
				log.Printf("Error reading message: %v", err)
				continue
			}

			// Process the message
			c.processMessage(ctx, msg)
		}
	}
}

// processMessage handles a single Kafka message
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	// Extract event type from headers
	var eventTypeStr string
	for _, header := range msg.Headers {
		if header.Key == "event_type" {
			eventTypeStr = string(header.Value)
			break
		}
	}

	if eventTypeStr == "" {
		log.Printf("Message missing event_type header, skipping")
		return
	}

	eventType := EventType(eventTypeStr)

	// Get handlers for this event type
	c.handlersLock.RLock()
	handlers := c.eventHandlers[eventType]
	c.handlersLock.RUnlock()

	if len(handlers) == 0 {
		// No handlers for this event type, commit message and continue
		return
	}

	// Process with all registered handlers
	for _, handler := range handlers {
		if err := handler(ctx, eventType, msg.Value); err != nil {
			log.Printf("Error handling event %s: %v", eventType, err)
			// Continue with other handlers rather than failing completely
		}
	}
}

// IsRunning returns whether the consumer is currently running
func (c *Consumer) IsRunning() bool {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	return c.running
}
