package query

import (
	"context"
	. "datastore/containers"
	. "datastore/utils"
	"go.uber.org/zap"
	"reflect"
	"sort"
	"sync"
)

const (
	dataTypeNotSupported   = "datatype not supported for aggregation"
	VerticalDayAggregators = 10
)

func GroupByVerticalAggregator(daysData []map[uint32][]DataPoint, aggregation string, queryTimeoutContext context.Context) {

	for dayIndex := 0; dayIndex < len(daysData); {

		select {

		case <-queryTimeoutContext.Done():

			return

		default:

			var completionWg sync.WaitGroup

			for range min(VerticalDayAggregators, len(daysData)-dayIndex) {

				if daysData[dayIndex] == nil {

					// Day Not present

					dayIndex++

					continue

				}

				completionWg.Add(1)

				go verticalAggregateSingleDay(daysData[dayIndex], aggregation, &completionWg)

				dayIndex++

			}

			completionWg.Wait()

		}

	}

}

func verticalAggregateSingleDay(day map[uint32][]DataPoint, aggregation string, completionWg *sync.WaitGroup) {

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
			aggregatedValue = Avg(batch)
		case "sum":
			aggregatedValue = Sum(batch)
		case "min":
			aggregatedValue = Min(batch)
		case "max":
			aggregatedValue = Max(batch)
		case "count":
			aggregatedValue = len(batch)
		default:
			Logger.Warn("aggregation not supported", zap.String("aggregation", aggregation), zap.Uint32("timestamp", timestamp), zap.Any("batch", batch))

		}

		day[0] = append(day[0], DataPoint{
			Timestamp: timestamp,

			Value: aggregatedValue,
		})
	}

}

func HorizontalAggregator(daysData []map[uint32][]DataPoint, aggregation string, interval uint32, from uint32, finalData map[uint32][]DataPoint, queryTimeoutContext context.Context) {

	objectWiseTimeIndexedBatchedData := make(map[uint32]map[uint32][]interface{})

	// Batching
	for _, day := range daysData {

		if day == nil {

			continue

		}

		for objectId, points := range day {

			select {

			case <-queryTimeoutContext.Done():

				return

			default:

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

	}

	// Reslice the days array, now object-wise data will be represented in single stream
	daysData = daysData[:]

	for objectId, timeIndexedBatch := range objectWiseTimeIndexedBatchedData {

		dataPoints := make([]DataPoint, 0)

		for timestamp, batch := range timeIndexedBatch {

			select {

			case <-queryTimeoutContext.Done():

				return

			default:

				var aggregatedValue interface{}

				switch aggregation {

				case "avg":
					aggregatedValue = Avg(batch)

				case "sum":
					aggregatedValue = Sum(batch)

				case "min":
					aggregatedValue = Min(batch)

				case "max":
					aggregatedValue = Max(batch)

				case "count":
					aggregatedValue = len(batch)

				default:
					Logger.Error("aggregation not supported", zap.String("aggregation", aggregation))

				}

				dataPoints = append(dataPoints, DataPoint{
					Timestamp: timestamp,
					Value:     aggregatedValue,
				})

			}

		}

		// Sort the final Data by timestamp
		sort.Slice(dataPoints, func(i, j int) bool {

			return dataPoints[i].Timestamp < dataPoints[j].Timestamp

		})

		finalData[objectId] = dataPoints

	}

}

func Max(values []interface{}) interface{} {

	switch dataType := reflect.TypeOf(values[0]).Kind(); dataType {

	case reflect.Float64:

		maxValue := values[0].(float64)

		for _, value := range values[1:] {

			if maxValue < value.(float64) {

				maxValue = value.(float64)

			}

		}

		return maxValue

	case reflect.Int64:

		maxValue := values[0].(int64)

		for _, value := range values[1:] {

			if maxValue < value.(int64) {

				maxValue = value.(int64)

			}

		}

		return maxValue

	default:

		Logger.Error(dataTypeNotSupported, zap.Any("datatype", dataType))

	}

	return nil
}

func Min(values []interface{}) interface{} {

	switch dataType := reflect.TypeOf(values[0]).Kind(); dataType {

	case reflect.Float64:

		minValue := values[0].(float64)

		for _, value := range values[1:] {

			if minValue > value.(float64) {

				minValue = value.(float64)

			}

		}

		return minValue

	case reflect.Int64:

		minValue := values[0].(int64)

		for _, value := range values[1:] {

			if minValue > value.(int64) {

				minValue = value.(int64)

			}

		}

		return minValue

	default:

		Logger.Error(dataTypeNotSupported, zap.Any("datatype", dataType))

	}

	return nil
}

func Sum(values []interface{}) interface{} {

	switch dataType := reflect.TypeOf(values[0]).Kind(); dataType {

	case reflect.Float64:
		sum := 0.0

		for _, value := range values {

			sum += value.(float64)

		}

		return sum

	case reflect.Int64:
		var sum int64 = 0

		for _, value := range values {

			sum += value.(int64)

		}

		return sum

	default:

		Logger.Error(dataTypeNotSupported, zap.Any("datatype", dataType))

	}

	return nil

}

func Avg(values []interface{}) interface{} {

	sum := Sum(values)

	switch dataType := reflect.TypeOf(values[0]).Kind(); dataType {

	case reflect.Float64:

		return sum.(float64) / float64(len(values))

	case reflect.Int64:

		return float64(sum.(int64)) / float64(len(values))

	default:

		Logger.Error(dataTypeNotSupported, zap.Any("datatype", dataType))

	}

	return nil

}
