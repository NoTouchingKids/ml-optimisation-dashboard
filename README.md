# ml-optimisation-dashboard

# Summary of the Overall Architecture

Let me summarize the redesigned event-driven architecture that maintains your existing technologies while providing better scalability and performance:

Note : This not just for machine learning but also Mathmatical optimisstion as well like guroby py

## Core Components

the philosophy is Backend should be Monolith and rest should suport functions should microservices which can be scaled. The backend worldn't be suturated as there limited no processing that need to be done in it.

### Go Backend (as Monolith when posible)

1. **API Gateway**

   - Entry point for client HTTP/HTTPS requests and WebSocket connections
   - Routes commands to appropriate services

2. **Command Service**

   - Processes user commands (train model, run prediction)
   - Validates requests and publishes events to the Event Backbone
   - Handles authorization and validation
     This Service also handle all client interactions like usermanagement and may be auth using JWT.

3. **ML Orchestrator**

   - Subscribes to command events from the Event Backbone
   - Communicates with Python ML services via gRPC
   - Tracks job status and publishes status updates to Event Backbone
   - Also tack Python ML services cluster's health satus and stating and stoping clusters.

4. **Query Service**

   - Maintains materialized views of system state
   - Provides APIs for historical data and analytics
   - Subscribes to events from Event Backbone for state updates

5. **Log Streaming Service**
   Not sure if this need to be part of the Monolith
   - Receives logs via UDP from Python services
   - Optimized for high-throughput (50-100+ logs per second per model)
   - Direct streaming to WebSocket clients without going through Event Backbone
     > Note: WebSockets used specifically because SSE doesn't support byte arrays
   - Stores all logs in TimescaleDB

### Event / message bus

- Central message bus for control messages and workflow orchestration
- Handles commands, status updates, and system events
- Implemented with a message broker Kafka

### Kafka

- Event backbone for Control Plane
- Handles commands, statuses, and workflow events
- Provides persistence and replay capabilities

### Python ML Service

This definatly a microservices and act as cluster were the main service would handle request from `ML Orchestrator` and spawn/flork worker to do ML or Optimistioan.

- Receives requests via gRPC from ML Orchestration Module
- Executes ML tasks (training, prediction)
- Sends logs via UDP to Log Streaming Module
- Returns results via gRPC

### TimescaleDB

- Stores time-series log data
- Optimized for time-based queries
- Maintains history for analytics

## Key Information Flows

### Command Flow

1. Client sends command via HTTP
2. API Gateway validates and publishes command event to Event Backbone
3. ML Orchestrator receives command event and calls Python service via gRPC
4. Status updates flow back through Event Backbone
5. WebSocket connections receive status updates

### Log Flow

1. Python service sends logs via UDP to Log Streaming Service
2. Log Streaming Service processes and stores logs in TimescaleDB
3. Log Streaming Service streams logs directly to connected WebSockets else if the client is not conneted that buffer n amount new loggs FIFO struture.
4. Clients receive logs in batches for efficient processing
