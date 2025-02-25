package handler

import (
	"backend/internal/database"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TrainingHandler struct {
	db *database.Client
}

// GET /api/training/data-sources
func (h *TrainingHandler) GetDataSources(c *gin.Context) {
	sources := []struct {
		Name string `json:"name"`
		Type string `json:"type"` // table/view/procedure
	}{
		// Query from information_schema
	}
	c.JSON(http.StatusOK, sources)
}

// GET /api/training/data
func (h *TrainingHandler) GetTrainingData(c *gin.Context) {
	source := c.Query("source")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return
	}

	data, total, err := h.db.GetTrainingData(source, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  data,
		"total": total,
		"page":  page,
	})
}
