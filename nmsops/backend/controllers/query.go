package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	. "nms-backend/db"
)

type QueryController struct {
	ReportDB *ReportDbClient
}

func NewQueryController(report *ReportDbClient) *QueryController {
	return &QueryController{
		ReportDB: report,
	}
}

func (queryController *QueryController) HandleQuery(ctx *gin.Context) {
	var req Query

	// Bind JSON from request body

	if err := ctx.ShouldBindJSON(&req); err != nil {

		log.Printf("JSON binding error: %v", err)

		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid JSON request: %v", err)})

		return

	}

	// Validate required fields

	if req.From == 0 || req.To == 0 || req.CounterId == 0 {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "from, to, and counterID are required"})

		return

	}

	// Validate ObjectIDs aren't empty

	if len(req.ObjectIds) == 0 {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "objectIDs must contain at least one value"})

		return

	}

	// Validate Aggregators

	switch req.VerticalAggregation {

	case "avg", "sum", "min", "max", "none":

	default:

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid aggregation function. Its must be either 'avg', 'sum', 'min', 'max' or 'none'"})

		return

	}

	switch req.HorizontalAggregation {

	case "avg", "sum", "min", "max", "none":

	default:

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid aggregation function. Its must be either 'avg', 'sum', 'min', 'max' or 'none'"})

		return

	}

	if req.Interval < 0 {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "interval cannot be negative"})

		return

	}

	// Log the parsed request for debugging
	log.Printf("Processed request: From=%d, To=%d, CounterID=%d, ObjectIDs=%v",
		req.From, req.To, req.CounterId, req.ObjectIds)

	response, err := queryController.ReportDB.Query(req.From, req.To, req.Interval, req.ObjectIds, req.CounterId, req.VerticalAggregation, req.HorizontalAggregation)

	if err != nil {

		log.Printf("Database query error: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Database error: %v", err)})

		return

	}

	ctx.JSON(http.StatusOK, response)
}
