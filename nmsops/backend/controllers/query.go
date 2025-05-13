package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	. "nms-backend/db"
	. "nms-backend/models"
	. "nms-backend/utils"
)

type QueryController struct {
	ReportDB *ReportDBClient
}

func NewQueryController(report *ReportDBClient) *QueryController {

	return &QueryController{

		ReportDB: report,
	}

}

func (queryController *QueryController) HandleQuery(ctx *gin.Context) {
	var req UserQueryRequest

	// Bind JSON from request body

	if err := ctx.ShouldBindJSON(&req); err != nil {

		Logger.Error("Error parsing request", zap.Error(err))

		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid JSON request: %v", err)})

		return

	}

	// Validate required fields

	if req.From == 0 || req.To == 0 || req.CounterId == 0 {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "from, to, and counterID are required"})

		return

	}

	// Validate the from and to range

	if req.From > req.To {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range, from must be less than or equal to to"})

		return
	}

	// Validate Aggregators

	switch req.ObjectWiseAggregation {

	case "avg", "sum", "min", "max", "count", "none":

	default:

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid aggregation function. Its must be either 'avg', 'sum', 'min', 'max', 'count' or 'none'"})

		return

	}

	switch req.TimestampAggregation {

	case "avg", "sum", "min", "max", "count", "none":

	default:

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid aggregation function. Its must be either 'avg', 'sum', 'min', 'max', 'count' or 'none'"})

		return

	}

	if req.Interval < 0 {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "interval cannot be negative"})

		return

	}

	response, err := queryController.ReportDB.Query(req.From, req.To, req.Interval, req.ObjectIds, req.CounterId, req.ObjectWiseAggregation, req.TimestampAggregation)

	if err != nil {

		Logger.Warn("Error querying database", zap.Error(err))

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: %v", err)})

		return

	}

	ctx.JSON(http.StatusOK, response)
}
