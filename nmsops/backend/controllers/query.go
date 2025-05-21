package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	. "nms-backend/db"
	. "nms-backend/utils"
	"time"
)

type userQueryRequest struct {
	From                  uint32   `json:"from" binding:"required"`
	To                    uint32   `json:"to" binding:"required"`
	ObjectIds             []string `json:"object_ids"`
	CounterId             uint16   `json:"counter_id" binding:"required"`
	ObjectWiseAggregation string   `json:"object_wise_aggregation" binding:"required"`
	TimestampAggregation  string   `json:"timestamp_aggregation" binding:"required"`
	Interval              uint32   `json:"interval"`
}

type QueryController struct {
	ReportDB *ReportDBClient
}

func NewQueryController(report *ReportDBClient) *QueryController {

	return &QueryController{

		ReportDB: report,
	}

}

func (queryController *QueryController) HandleQuery(ctx *gin.Context) {

	startTime := time.Now()

	var req userQueryRequest

	// Bind JSON from request body

	if err := ctx.ShouldBindJSON(&req); err != nil {

		Logger.Error("Error parsing request", zap.Error(err))

		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid JSON request: %v", err)})

		return

	}

	// ------------------- Validate Request Body --------------------

	// Validate the from and to range

	if req.From > req.To {

		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time range, from must be less than or equal to to"})

		return
	}

	// Validate ObjectIds
	for _, ip := range req.ObjectIds {

		if !ValidateIpAddress(ip) {

			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid IP address in object_ids"})

			return

		}

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

	// ------------------- Query ReportDB --------------------

	response, err := queryController.ReportDB.Query(req.From, req.To, req.Interval, req.ObjectIds, req.CounterId, req.ObjectWiseAggregation, req.TimestampAggregation)

	if err != nil {

		Logger.Warn("Error querying database", zap.Error(err))

		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error: %v", err)})

		return

	}

	ctx.JSON(http.StatusOK, response)

	Logger.Debug("Query executed", zap.Duration("duration", time.Since(startTime)))
}
