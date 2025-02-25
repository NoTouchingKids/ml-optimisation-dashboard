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
	"backend/internal/grpc"
	"backend/internal/handler"
	"backend/internal/store"

	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg        *config.Config
	router     *gin.Engine
	db         *database.Client
	userDB     *database.UserDB
	userStore  *store.UserStore
	jwtService *auth.JWTService
	grpcClient *grpc.Client
	logBuffer  *buffer.LogBuffer
}

func New(cfg *config.Config, tsdb *database.Client, userDB *database.UserDB,
	jwtService *auth.JWTService, userStore *store.UserStore) (*Server, error) {
	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("initializing database: %w", err)
	}

	// Initialize gRPC client
	logBuffer := buffer.NewLogBuffer(100) // Buffer 100 logs per client
	grpcClient, err := grpc.NewClient(cfg.GRPC.ServerAddress, logBuffer)
	if err != nil {
		return nil, fmt.Errorf("initializing gRPC client: %w", err)
	}

	// Initialize router
	router := gin.Default()

	server := &Server{
		cfg:        cfg,
		router:     router,
		db:         db,
		userDB:     userDB,
		userStore:  userStore,
		jwtService: jwtService,
		grpcClient: grpcClient,
		logBuffer:  logBuffer,
	}

	server.setupRoutes()
	return server, nil
}

func (s *Server) setupRoutes() {
	// Create handlers
	// logBuffer := buffer.NewLogBuffer(100)
	wsHandler := handler.NewWebSocketHandler(s.db, s.grpcClient, s.logBuffer)
	restHandler := handler.NewRESTHandler(s.db, s.grpcClient)

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
		api.POST("/model/train", restHandler.HandleTrain)
		api.POST("/model/predict", restHandler.HandlePredict)
		api.GET("/model/status/:clientId", restHandler.HandleStatus)
	}
}

func (s *Server) Run(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	log.Printf("Starting server on %s", addr)

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

// Modify Shutdown method to return an error
func (s *Server) Shutdown(ctx context.Context) error {
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
