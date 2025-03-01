# Ml & optimisation dashboard

TBC

## Summary of the Overall Architecture

### System Overview

The architecture consists of a monolithic Go backend and a separate ML Orchestrator microservice. This hybrid approach provides the simplicity of a monolith for primary application functions while isolating machine learning orchestration in a dedicated microservice for better scalability and separation of concerns.

### Key Components

#### 1. Monolithic Go Backend

- **Primary Application**: Serves as the core application and single entry point for all client interactions
- **Clean Architecture**: Implemented in layers (Domain, Application, Infrastructure, Presentation)
- **Client Communication**: Handles all REST API requests and WebSocket connections
- **Database Access**: Interacts with databases (TimescaleDB, PostgreSQL) but does not manage migrations
- **Event Publishing**: Publishes ML-related commands and status updates to Kafka

#### 2. ML Orchestrator Microservice

- **Event-Driven Service**: Listens to commands on the Kafka event bus
- **ML Workflow Management**: Orchestrates machine learning processes
- **Service Communication**: Communicates directly with Python ML services via gRPC
- **Resilience Patterns**: Implements circuit breaking, retries, and timeouts
- **Status Updates**: Publishes process updates back to Kafka for client notification

#### 3. Event Bus (Kafka)

- **Command Channel**: For ML process commands (train, predict)
- **Status Channel**: For ML process status updates
- **Internal Communication**: Used exclusively for communication between Go Backend and ML Orchestrator
- **Durable Messaging**: Ensures reliable command and event delivery

#### 4. Python ML Services

- **Model Processing**: Executes actual ML model training and inference
- **gRPC Server**: Exposes ML functionality via gRPC interfaces
- **Isolated Processing**: No direct access to clients or database

#### 5. Databases

- **TimescaleDB**: For time-series data like logs and metrics
- **PostgreSQL**: For relational data like users and model metadata
- **External Management**: Schema and migrations managed outside the Go application

### Data Flow

1. **Client Request Flow**:

   - Web client sends requests to Go Backend via REST API
   - Go Backend processes business logic and publishes commands to Kafka
   - ML Orchestrator consumes commands and communicates with Python ML services
   - Status updates flow back through Kafka to Go Backend
   - Go Backend notifies clients via WebSockets

2. **Log Processing Flow**:
   - ML services generate logs during processing
   - Logs are streamed via UDP to the Go Backend
   - Go Backend processes and persists logs to TimescaleDB
   - Real-time logs are streamed to clients via WebSockets

### Technology Stack

1. **Go Backend**:

   - Language: Go
   - Web Framework: Gin
   - Database Access: Standard library or sqlc
   - WebSockets: gorilla/websocket
   - Kafka Client: segmentio/kafka-go
   - Observability: zap or zerolog for logging, Prometheus for metrics

2. **ML Orchestrator**:

   - Language: Go
   - Kafka Client: segmentio/kafka-go
   - gRPC: google.golang.org/grpc
   - Resilience: circuitbreaker patterns, timeout middleware

3. **External Components**:
   - Kafka: Event streaming platform
   - TimescaleDB: Time-series database for logs and metrics
   - PostgreSQL: Relational database for user and model data
   - Python ML Services: gRPC servers running ML models
   - Mnio: for S3 like file store

### Key Design Principles

1. **Separation of Concerns**: Clean boundaries between components and layers
2. **Single Responsibility**: Each component has a clear, focused purpose
3. **Interface-Based Design**: Dependencies are defined through interfaces
4. **Observability**: Comprehensive logging, metrics, and tracing
5. **Resilience**: Fault tolerance through circuit breakers, retries, and timeouts
6. **Event-Driven Communication**: Loose coupling through event-based messaging

This architecture provides a robust foundation for ML workflows while maintaining clean separation between client-facing functionality and complex ML orchestration processes.
