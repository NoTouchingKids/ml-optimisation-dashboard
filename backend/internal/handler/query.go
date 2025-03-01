// backend/internal/handler/query.go
package handler

import (
	"net/http"
	"strconv"
	"time"

	"backend/internal/query"

	"github.com/gin-gonic/gin"
)

// QueryHandler handles queries for model and log data
type QueryHandler struct {
	queryService *query.QueryService
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(queryService *query.QueryService) *QueryHandler {
	return &QueryHandler{
		queryService: queryService,
	}
}

// GetModelState returns the current state of a specific model
func (h *QueryHandler) GetModelState(c *gin.Context) {
	clientID := c.Param("clientId")

	state, exists := h.queryService.GetModelState(clientID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Model not found"})
		return
	}

	c.JSON(http.StatusOK, state)
}

// GetRunningModels returns all currently running models
func (h *QueryHandler) GetRunningModels(c *gin.Context) {
	models := h.queryService.GetRunningModels()
	c.JSON(http.StatusOK, gin.H{"models": models, "count": len(models)})
}

// QueryModelHistory returns filtered model history
func (h *QueryHandler) QueryModelHistory(c *gin.Context) {
	// Parse query parameters
	clientID := c.Query("client_id")
	processType := c.Query("process_type")
	status := c.Query("status")

	// Parse time range parameters
	var fromTime, toTime time.Time
	var err error

	if fromStr := c.Query("from"); fromStr != "" {
		fromTime, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' time format"})
			return
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		toTime, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' time format"})
			return
		}
	}

	// Parse pagination parameters
	limit := 50 // Default limit
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
			return
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
			return
		}
	}

	// Create filter
	filter := query.QueryFilter{
		ClientID:      clientID,
		ProcessType:   processType,
		Status:        status,
		StartTimeFrom: fromTime,
		StartTimeTo:   toTime,
		Limit:         limit,
		Offset:        offset,
	}

	// Query model history
	history := h.queryService.QueryModelHistory(filter)

	c.JSON(http.StatusOK, gin.H{
		"models": history,
		"count":  len(history),
		"filter": filter,
	})
}

// GetLogSummary returns summarized log information
func (h *QueryHandler) GetLogSummary(c *gin.Context) {
	clientID := c.Param("clientId")

	// Parse time range parameters
	var fromTime, toTime time.Time
	var err error

	if fromStr := c.Query("from"); fromStr != "" {
		fromTime, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'from' time format"})
			return
		}
	} else {
		// Default to 24 hours ago
		fromTime = time.Now().Add(-24 * time.Hour)
	}

	if toStr := c.Query("to"); toStr != "" {
		toTime, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'to' time format"})
			return
		}
	} else {
		// Default to now
		toTime = time.Now()
	}

	// Get log summary
	summary, err := h.queryService.GetLogSummary(c.Request.Context(), clientID, fromTime, toTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
