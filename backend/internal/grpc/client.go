package grpc

import (
	"backend/internal/buffer"
	pb "backend/proto"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn      *grpc.ClientConn
	client    pb.ProcessServiceClient
	stream    pb.ProcessService_StreamLogsClient
	logBuffer *buffer.LogBuffer
}

func NewClient(serverAddr string, logBuffer *buffer.LogBuffer) (*Client, error) {
	// dialCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	client := &Client{
		conn:      conn,
		client:    pb.NewProcessServiceClient(conn),
		logBuffer: logBuffer,
	}

	if err := client.setupStreamConnection(); err != nil {
		conn.Close()
		return nil, err
	}

	go client.handleLogStream()
	return client, nil
}

func (c *Client) setupStreamConnection() error {
	stream, err := c.client.StreamLogs(context.Background(), &pb.LogRequest{ClientId: "3a44390c-c7b6-43b9-9cdc-1dcc9bb7d794"})
	if err != nil {
		return fmt.Errorf("failed to setup stream: %w", err)
	}
	c.stream = stream
	return nil
}

func (c *Client) handleLogStream() {
	for {
		msg, err := c.stream.Recv()
		if err != nil {
			log.Printf("Stream error: %v, attempting reconnect...", err)
			time.Sleep(1 * time.Second)
			if err := c.setupStreamConnection(); err != nil {
				log.Printf("Reconnection failed: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}
			continue
		}
		// log.Printf("Received log: %+v", msg)
		// Store received log in buffer
		record := buffer.LogRecord{
			Timestamp: msg.Timestamp,
			ClientID:  msg.ClientId,
			Message:   msg.Message,
			ProcessID: msg.ProcessId,
		}
		c.logBuffer.Push(msg.ClientId, record)
		log.Printf("Buffered log for client %s", msg.ClientId)

	}
}

func (c *Client) StartProcess(ctx context.Context, clientID string, payload []byte) (*pb.ProcessResponse, error) {
	req := &pb.StartProcessRequest{
		ClientId: clientID,
		Payload:  string(payload),
	}

	return c.client.StartProcess(ctx, req)
}

func (c *Client) Close() error {
	if c.stream != nil {
		c.stream.CloseSend()
	}
	return c.conn.Close()
}
