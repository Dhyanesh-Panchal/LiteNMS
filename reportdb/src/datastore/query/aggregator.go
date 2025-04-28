package query

import (
	. "datastore/containers"
	. "datastore/utils"
	"go.uber.org/zap"
	"sort"
	"sync"
)

const (
	dataTypeNotSupported = "datatype not supported for aggregation"
)

func GroupByVerticalAggregator(daysData []map[uint32][]DataPoint, aggregation string, dataType string) {

	var completionWg sync.WaitGroup

	for index := range len(daysData) {

		completionWg.Add(1)

		go verticalAggregateSingleDay(daysData[index], aggregation, dataType, &completionWg)

	}

	completionWg.Wait()

}

func verticalAggregateSingleDay(day map[uint32][]DataPoint, aggregation string, dataType string, completionWg *sync.WaitGroup) {

	defer completionWg.Done()

	timeIndexedBatchedData := make(map[uint32][]interface{})

	for objectId, points := range day {

		for _, point := range points {

			timeIndexedBatchedData[point.Timestamp] = append(timeIndexedBatchedData[point.Timestamp], point.Value)

		}

		delete(day, objectId)

	}

	// Using 0 as the final objectId which is aggregate of all the objectIds
	day[0] = make([]DataPoint, 0)

	for timestamp, batch := range timeIndexedBatchedData {

		var aggregatedValue interface{}

		switch aggregation {
		case "avg":
			aggregatedValue = Avg(batch, dataType)
		case "sum":
			aggregatedValue = Sum(batch, dataType)
		case "min":
			aggregatedValue = Min(batch, dataType)
		case "max":
			aggregatedValue = Max(batch, dataType)
		default:
			Logger.Warn("aggregation not supported", zap.String("aggregation", aggregation), zap.Uint32("timestamp", timestamp), zap.Any("batch", batch), zap.String("dataType", dataType))

		}

		day[0] = append(day[0], DataPoint{
			Timestamp: timestamp,

			Value: aggregatedValue,
		})
	}

}

func HorizontalAggregator(daysData []map[uint32][]DataPoint, aggregation string, dataType string, interval uint32, from uint32) map[uint32][]DataPoint {

	objectWiseTimeIndexedBatchedData := make(map[uint32]map[uint32][]interface{})

	// Batching
	for _, day := range daysData {

		for objectId, points := range day {

			if _, exist := objectWiseTimeIndexedBatchedData[objectId]; !exist {

				objectWiseTimeIndexedBatchedData[objectId] = make(map[uint32][]interface{})

			}

			for _, point := range points {

				if interval != 0 {

					// Histogram Interval
					currentTimestamp := point.Timestamp - from // normalizing the time range to start histogram interval at 'from' timestamp.

					histogramTimestamp := (currentTimestamp - currentTimestamp%interval) + from

					objectWiseTimeIndexedBatchedData[objectId][histogramTimestamp] = append(objectWiseTimeIndexedBatchedData[objectId][histogramTimestamp], point.Value)
				} else {

					// Gauge/Grid aggregation, No interval present
					// Using 0 as the common timestamp to aggregate whole from-to range

					objectWiseTimeIndexedBatchedData[objectId][0] = append(objectWiseTimeIndexedBatchedData[objectId][0], point.Value)

				}

			}

		}

	}

	// Reslice the days array, now object-wise data will be represented in single stream
	daysData = daysData[:]

	// Aggregation over the batch
	finalData := make(map[uint32][]DataPoint)

	for objectId, timeIndexedBatch := range objectWiseTimeIndexedBatchedData {

		dataPoints := make([]DataPoint, 0)

		for timestamp, batch := range timeIndexedBatch {

			var aggregatedValue interface{}

			switch aggregation {

			case "avg":
				aggregatedValue = Avg(batch, dataType)

			case "sum":
				aggregatedValue = Sum(batch, dataType)

			case "min":
				aggregatedValue = Min(batch, dataType)

			case "max":
				aggregatedValue = Max(batch, dataType)

			default:
				Logger.Error("aggregation not supported", zap.String("aggregation", aggregation))

			}

			dataPoints = append(dataPoints, DataPoint{
				Timestamp: timestamp,
				Value:     aggregatedValue,
			})
		}

		// Sort the final Data by timestamp
		sort.Slice(dataPoints, func(i, j int) bool {

			return dataPoints[i].Timestamp < dataPoints[j].Timestamp

		})

		finalData[objectId] = dataPoints

	}

	return finalData
}

func Max(values []interface{}, dataType string) interface{} {

	switch dataType {

	case "float64":
		maxValue := values[0].(float64)

		for _, value := range values[1:] {

			if maxValue < value.(float64) {

				maxValue = value.(float64)

			}
		}

		return maxValue

	case "int64":
		maxValue := values[0].(int64)

		for _, value := range values[1:] {

			if maxValue < value.(int64) {

				maxValue = value.(int64)

			}
		}

		return maxValue

	}

	Logger.Error(dataTypeNotSupported, zap.String("datatype", dataType))

	return nil

}

func Min(values []interface{}, dataType string) interface{} {

	switch dataType {

	case "float64":
		minValue := values[0].(float64)

		for _, value := range values[1:] {

			if minValue > value.(float64) {

				minValue = value.(float64)

			}
		}

		return minValue

	case "int64":
		minValue := values[0].(int64)

		for _, value := range values[1:] {

			if minValue > value.(int64) {

				minValue = value.(int64)

			}
		}

		return minValue

	}

	Logger.Error(dataTypeNotSupported, zap.String("datatype", dataType))

	return nil

}

func Sum(values []interface{}, dataType string) interface{} {

	switch dataType {

	case "float64":
		sum := 0.0

		for _, value := range values {

			sum += value.(float64)

		}

		return sum

	case "int64":
		var sum int64 = 0

		for _, value := range values {

			sum += value.(int64)

		}

		return sum

	}

	Logger.Error(dataTypeNotSupported, zap.String("datatype", dataType))

	return nil

}

func Avg(values []interface{}, dataType string) interface{} {

	sum := Sum(values, dataType)

	switch dataType {

	case "float64":

		return sum.(float64) / float64(len(values))

	case "int64":
		return float64(sum.(int64)) / float64(len(values))

	}

	Logger.Error(dataTypeNotSupported, zap.String("datatype", dataType))

	return nil

}
