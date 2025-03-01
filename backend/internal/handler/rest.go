package handler

import (
	"fmt"
	"log"
	"net/http"

	"backend/internal/database"
	"backend/internal/event"
	"backend/internal/grpc"
	"backend/internal/types"

	"github.com/gin-gonic/gin"
)

type RESTHandler struct {
	db         *database.Client
	grpcClient *grpc.Client
	producer   *event.Producer
}

func NewRESTHandler(db *database.Client, grpcClient *grpc.Client, producer *event.Producer) *RESTHandler {
	return &RESTHandler{
		db:         db,
		grpcClient: grpcClient,
		producer:   producer,
	}
}

func (h *RESTHandler) HandleTrain(c *gin.Context) {
	var req types.ModelRequest
	if err := c.BindJSON(&req); err != nil {
		log.Printf("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Publish train request event to Kafka
	err := h.producer.PublishTrainRequest(
		c.Request.Context(),
		req.ClientID,
		req.Data,
		req.StartDate,
		req.EndDate,
		req.Configuration,
	)
	if err != nil {
		log.Printf("Failed to publish train event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to publish event: %v", err)})
		return
	}

	// Return immediate acknowledgment
	c.JSON(http.StatusAccepted, gin.H{
		"client_id": req.ClientID,
		"status":    "pending",
		"message":   "Training request has been queued",
	})
}

func (h *RESTHandler) HandlePredict(c *gin.Context) {
	var req types.ModelRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Publish predict request event to Kafka
	err := h.producer.PublishPredictRequest(
		c.Request.Context(),
		req.ClientID,
		req.Data,
		req.Configuration,
	)
	if err != nil {
		log.Printf("Failed to publish predict event: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to publish event: %v", err)})
		return
	}

	// Return immediate acknowledgment
	c.JSON(http.StatusAccepted, gin.H{
		"client_id": req.ClientID,
		"status":    "pending",
		"message":   "Prediction request has been queued",
	})
}

func (h *RESTHandler) HandleStatus(c *gin.Context) {
	clientID := c.Param("clientId")
	status, err := h.db.GetModelStatus(c.Request.Context(), clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}
