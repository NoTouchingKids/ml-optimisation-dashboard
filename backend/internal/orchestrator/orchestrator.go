package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/event"
	"backend/internal/grpc"
	"backend/internal/types"
)

// MLOrchestrator manages ML processes and status updates
type MLOrchestrator struct {
	grpcClient *grpc.Client
	producer   *event.Producer
	consumer   *event.Consumer
}

// NewMLOrchestrator creates a new ML Orchestrator
func NewMLOrchestrator(grpcClient *grpc.Client, producer *event.Producer, consumer *event.Consumer) *MLOrchestrator {
	orchestrator := &MLOrchestrator{
		grpcClient: grpcClient,
		producer:   producer,
		consumer:   consumer,
	}

	// Subscribe to command events
	consumer.Subscribe(event.EventTypeTrainRequested, orchestrator.handleTrainRequest)
	consumer.Subscribe(event.EventTypePredictRequested, orchestrator.handlePredictRequest)

	return orchestrator
}

// Start begins the orchestrator operation
func (o *MLOrchestrator) Start(ctx context.Context) {
	o.consumer.Start(ctx)
}

// Stop halts the orchestrator operation
func (o *MLOrchestrator) Stop() {
	o.consumer.Stop()
}

// handleTrainRequest processes a training request
func (o *MLOrchestrator) handleTrainRequest(ctx context.Context, eventType event.EventType, data []byte) error {
	// Deserialize event
	e, err := event.Deserialize(data, eventType)
	if err != nil {
		return fmt.Errorf("deserializing train event: %w", err)
	}

	trainEvent, ok := e.(*event.TrainRequestedEvent)
	if !ok {
		return fmt.Errorf("expected TrainRequestedEvent but got %T", e)
	}

	// Create the request for the gRPC service
	modelReq := types.ModelRequest{
		Type:          "train",
		ClientID:      trainEvent.ClientID,
		Data:          trainEvent.Data,
		StartDate:     trainEvent.StartDate,
		EndDate:       trainEvent.EndDate,
		Configuration: trainEvent.Configuration,
	}

	reqBytes, err := json.Marshal(modelReq)
	if err != nil {
		return fmt.Errorf("marshaling model request: %w", err)
	}

	// Call the Python ML service via gRPC
	resp, err := o.grpcClient.StartProcess(ctx, trainEvent.ClientID, reqBytes)
	if err != nil {
		// Publish failure event
		o.producer.PublishModelStatus(
			ctx,
			event.EventTypeModelFailed,
			trainEvent.ClientID,
			"error",
			fmt.Sprintf("Failed to start training: %v", err),
			"train",
			0,
		)
		return fmt.Errorf("starting training process: %w", err)
	}

	// Publish started event
	return o.producer.PublishModelStatus(
		ctx,
		event.EventTypeModelStarted,
		trainEvent.ClientID,
		resp.Status,
		"Training process started",
		"train",
		0,
	)
}

// handlePredictRequest processes a prediction request
func (o *MLOrchestrator) handlePredictRequest(ctx context.Context, eventType event.EventType, data []byte) error {
	// Deserialize event
	e, err := event.Deserialize(data, eventType)
	if err != nil {
		return fmt.Errorf("deserializing predict event: %w", err)
	}

	predictEvent, ok := e.(*event.PredictRequestedEvent)
	if !ok {
		return fmt.Errorf("expected PredictRequestedEvent but got %T", e)
	}

	// Create the request for the gRPC service
	modelReq := types.ModelRequest{
		Type:          "predict",
		ClientID:      predictEvent.ClientID,
		Data:          predictEvent.Data,
		Configuration: predictEvent.Configuration,
	}

	reqBytes, err := json.Marshal(modelReq)
	if err != nil {
		return fmt.Errorf("marshaling model request: %w", err)
	}

	// Call the Python ML service via gRPC
	resp, err := o.grpcClient.StartProcess(ctx, predictEvent.ClientID, reqBytes)
	if err != nil {
		// Publish failure event
		o.producer.PublishModelStatus(
			ctx,
			event.EventTypeModelFailed,
			predictEvent.ClientID,
			"error",
			fmt.Sprintf("Failed to start prediction: %v", err),
			"predict",
			0,
		)
		return fmt.Errorf("starting prediction process: %w", err)
	}

	// Publish started event
	return o.producer.PublishModelStatus(
		ctx,
		event.EventTypeModelStarted,
		predictEvent.ClientID,
		resp.Status,
		"Prediction process started",
		"predict",
		0,
	)
}

// ProcessLogToStatus processes log data to extract status updates
func (o *MLOrchestrator) ProcessLogToStatus(ctx context.Context, log types.LogRecord) error {
	// In a real implementation, you would parse the log message to see if it contains
	// status update information, and if so, publish a status event

	// For now, let's keep this as a placeholder for future implementation
	return nil
}
