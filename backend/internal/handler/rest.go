package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"backend/internal/database"
	"backend/internal/grpc"
	"backend/internal/types"

	"github.com/gin-gonic/gin"
)

type RESTHandler struct {
	db         *database.Client
	grpcClient *grpc.Client
}

func NewRESTHandler(db *database.Client, grpcClient *grpc.Client) *RESTHandler {
	return &RESTHandler{
		db:         db,
		grpcClient: grpcClient,
	}
}

func (h *RESTHandler) HandleTrain(c *gin.Context) {
	var req types.ModelRequest
	if err := c.BindJSON(&req); err != nil {
		log.Printf("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Log the payload being sent
	// log.Printf("Sending payload to gRPC: %+v", payload)

	resp, err := h.grpcClient.StartProcess(c.Request.Context(), req.ClientID, payload)
	if err != nil {
		log.Printf("gRPC call failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("gRPC error: %v", err)})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *RESTHandler) HandlePredict(c *gin.Context) {
	var req types.ModelRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.grpcClient.StartProcess(c.Request.Context(), req.ClientID, payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
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
