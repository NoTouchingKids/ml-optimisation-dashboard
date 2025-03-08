// backend/internal/server/server.go
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"backend/internal/auth"
	"backend/internal/buffer"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/event"
	"backend/internal/grpc"
	"backend/internal/handler"
	"backend/internal/orchestrator"
	"backend/internal/query"
	"backend/internal/store"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg             *config.Config
	router          *gin.Engine
	db              *database.Client
	userDB          *database.UserDB
	userStore       *store.UserStore
	jwtService      *auth.JWTService
	grpcClient      *grpc.Client
	logBuffer       *buffer.LogBuffer
	producer        *event.Producer
	commandConsumer *event.Consumer
	statusConsumer  *event.Consumer
	orchestrator    *orchestrator.MLOrchestrator
	statusHandler   *handler.StatusHandler
	queryService    *query.QueryService
}

func New(cfg *config.Config, tsdb *database.Client, userDB *database.UserDB,
	jwtService *auth.JWTService, userStore *store.UserStore) (*Server, error) {
	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("initializing database: %w", err)
	}

	// Initialize log buffer
	logBuffer := buffer.NewLogBuffer(100) // Buffer 100 logs per client

	// Initialize gRPC client
	grpcClient, err := grpc.NewClient(cfg.GRPC.ServerAddress, logBuffer)
	if err != nil {
		return nil, fmt.Errorf("initializing gRPC client: %w", err)
	}

	// Initialize Kafka producers/consumers
	producer := event.NewProducer(
		cfg.Kafka.Brokers,
		cfg.Kafka.CommandTopic,
		cfg.Kafka.StatusTopic,
	)

	// Create separate consumers for different components
	commandConsumer := event.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.ConsumerGroup+"-command",
		[]string{cfg.Kafka.CommandTopic},
	)

	statusConsumer := event.NewConsumer(
		cfg.Kafka.Brokers,
		cfg.Kafka.ConsumerGroup+"-status",
		[]string{cfg.Kafka.StatusTopic},
	)

	// Initialize router
	router := gin.Default()

	// Setup WebSocket handler
	wsHandler := handler.NewWebSocketHandler(db, grpcClient, logBuffer)

	// Setup Status handler
	statusHandler := handler.NewStatusHandler(db, wsHandler, statusConsumer)

	// Setup ML Orchestrator
	mlOrchestrator := orchestrator.NewMLOrchestrator(grpcClient, producer, commandConsumer)

	if err != nil {
		return nil, fmt.Errorf("initializing log streaming service: %w", err)
	}

	// Setup Query Service
	queryService := query.NewQueryService(db, statusConsumer)

	server := &Server{
		cfg:             cfg,
		router:          router,
		db:              db,
		userDB:          userDB,
		userStore:       userStore,
		jwtService:      jwtService,
		grpcClient:      grpcClient,
		logBuffer:       logBuffer,
		producer:        producer,
		commandConsumer: commandConsumer,
		statusConsumer:  statusConsumer,
		orchestrator:    mlOrchestrator,
		statusHandler:   statusHandler,
		queryService:    queryService,
	}

	server.setupRoutes(wsHandler)
	return server, nil
}

func (s *Server) setupRoutes(wsHandler *handler.WebSocketHandler) {
	// Create handlers
	restHandler := handler.NewRESTHandler(s.db, s.grpcClient, s.producer)
	queryHandler := handler.NewQueryHandler(s.queryService)

	// CORS middleware
	s.router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// WebSocket route
	s.router.GET("/ws", wsHandler.HandleConnection)

	// REST routes
	api := s.router.Group("/api")
	{
		// Command routes
		api.POST("/model/train", restHandler.HandleTrain)
		api.POST("/model/predict", restHandler.HandlePredict)
		api.GET("/model/status/:clientId", restHandler.HandleStatus)

		// Query routes
		query := api.Group("/query")
		{
			query.GET("/model/:clientId", queryHandler.GetModelState)
			query.GET("/models/running", queryHandler.GetRunningModels)
			query.GET("/models/history", queryHandler.QueryModelHistory)
			query.GET("/logs/:clientId/summary", queryHandler.GetLogSummary)
		}
	}
}

func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	log.Printf("Starting server on %s", addr)

	// Start the Log Streaming Service
	// if err := s.logStreamService.Start(ctx); err != nil {
	// 	return fmt.Errorf("starting log streaming service: %w", err)
	// }

	// Start the Query Service
	if err := s.queryService.Start(ctx); err != nil {
		return fmt.Errorf("starting query service: %w", err)
	}

	// Start the ML orchestrator
	s.orchestrator.Start(ctx)

	// Start the status handler
	s.statusHandler.Start(ctx)

	srv := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Stop the Log Streaming Service
	// if err := s.logStreamService.Stop(); err != nil {
	// 	log.Printf("Log streaming service shutdown error: %v", err)
	// }

	// Stop the Query Service
	s.queryService.Stop()

	// Stop the ML orchestrator
	s.orchestrator.Stop()

	// Stop the status handler
	s.statusHandler.Stop()

	// Close the Kafka consumers
	s.commandConsumer.Stop()
	s.statusConsumer.Stop()

	// Close the Kafka producer
	if err := s.producer.Close(); err != nil {
		log.Printf("Kafka producer shutdown error: %v", err)
	}

	if err := s.db.Close(); err != nil {
		log.Printf("Database shutdown error: %v", err)
		return err
	}
	if err := s.grpcClient.Close(); err != nil {
		log.Printf("gRPC client shutdown error: %v", err)
		return err
	}
	return nil
}
