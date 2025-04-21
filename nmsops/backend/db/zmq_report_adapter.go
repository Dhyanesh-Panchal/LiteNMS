package db

import (
	"nms-backend/models"
	"nms-backend/reportdb"
)

// ZMQReportAdapter adapts the ReportDbClient to the ReportDB interface
type ZMQReportAdapter struct {
	client *reportdb.ReportDbClient
}

// NewZMQReportAdapter creates a new ZMQReportAdapter
func NewZMQReportAdapter(client *reportdb.ReportDbClient) ReportDB {
	return &ZMQReportAdapter{
		client: client,
	}
}

// QueryHistogram implements the ReportDB interface
func (adapter *ZMQReportAdapter) QueryHistogram(from, to int64, counterID uint16, objectIDs []uint32) (map[uint32][]models.HistogramPoint, error) {
	// Convert int64 timestamps to uint32 as required by reportdb client
	fromU32 := uint32(from)

	toU32 := uint32(to)

	// Query the report database
	results, err := adapter.client.Query(fromU32, toU32, objectIDs, counterID)

	if err != nil {

		return nil, err

	}

	// Convert DataPoint to HistogramPoint
	convertedResults := make(map[uint32][]models.HistogramPoint)

	for objID, points := range results {

		histPoints := make([]models.HistogramPoint, len(points))

		for i, point := range points {

			// Convert uint32 timestamp to int64
			timestamp := int64(point.Timestamp)

			// Extract value and convert to float64
			value := point.Value.(float64)

			//switch v := point.Value.(type) {
			//
			//case float64:
			//	value = v
			//
			//case float32:
			//	value = float64(v)
			//
			//case int:
			//	value = float64(v)
			//
			//case int64:
			//	value = float64(v)
			//
			//case uint:
			//	value = float64(v)
			//
			//case uint64:
			//	value = float64(v)
			//
			//default:
			//	// Default to 0 if conversion not possible
			//	value = 0
			//
			//}

			histPoints[i] = models.HistogramPoint{
				Timestamp: timestamp,
				Value:     value,
			}
		}
		convertedResults[objID] = histPoints
	}

	return convertedResults, nil
}
