package controllers

import (
	"fmt"
	"log"
	"net/http"
	"nms-backend/db"
	"nms-backend/models"

	"github.com/gin-gonic/gin"
)

type HistogramController struct {
	ReportDB db.ReportDB
}

func NewHistogramController(report db.ReportDB) *HistogramController {
	return &HistogramController{
		ReportDB: report,
	}
}

func (ctrl *HistogramController) GetHistogram(c *gin.Context) {
	var req models.HistogramQueryRequest
	
	// Bind JSON from request body
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("JSON binding error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid JSON request: %v", err)})
		return
	}

	// Validate required fields
	if req.From == 0 || req.To == 0 || req.CounterID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from, to, and counterID are required"})
		return
	}

	// Validate ObjectIDs aren't empty
	if len(req.ObjectIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "objectIDs must contain at least one value"})
		return
	}

	// Log the parsed request for debugging
	log.Printf("Processed request: From=%d, To=%d, CounterID=%d, ObjectIDs=%v", 
		req.From, req.To, req.CounterID, req.ObjectIDs)

	points, err := ctrl.ReportDB.QueryHistogram(req.From, req.To, req.CounterID, req.ObjectIDs)
	if err != nil {
		log.Printf("Database query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", err)})
		return
	}

	// Convert map[uint32][] → map[int][]
	converted := make(map[int][]models.HistogramPoint)
	for k, v := range points {
		converted[int(k)] = v
	}

	response := models.HistogramResponse{Data: converted}
	c.JSON(http.StatusOK, response)
}
