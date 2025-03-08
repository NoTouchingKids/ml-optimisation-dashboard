package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"backend/internal/auth"
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/server"
	"backend/internal/store"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Load configuration
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize TimescaleDB
	tsdb, err := database.New(
		config.DatabaseConfig{
			Host:     cfg.Database.Host,
			Port:     cfg.Database.Port,
			User:     cfg.Database.User,
			Password: cfg.Database.Password,
			DBName:   cfg.Database.DBName,
		},
	)

	if err != nil {
		log.Fatalf("Failed to initialize TimescaleDB: %v", err)
	}
	defer tsdb.Close()

	// Initialize User Database (PostgreSQL)
	userDB, err := database.NewUserDB(config.DatabaseConfig{
		Host:     os.Getenv("USER_DB_HOST"),
		Port:     5433, // User DB port
		User:     os.Getenv("USER_DB_USER"),
		Password: os.Getenv("USER_DB_PASSWORD"),
		DBName:   os.Getenv("USER_DB_NAME"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize User DB: %v", err)
	}
	defer userDB.Close()

	// Initialize JWT Service
	jwtService := auth.NewJWTService(os.Getenv("JWT_SECRET"))

	// Initialize User Store
	userStore := store.NewUserStore(userDB)

	// Initialize server with both databases and JWT service
	srv, err := server.New(cfg, tsdb, userDB, jwtService, userStore)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		cancel()
	}()

	// Run server
	if err := srv.Run(ctx); err != nil {
		log.Printf("Server error: %v", err)
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
	}
}
