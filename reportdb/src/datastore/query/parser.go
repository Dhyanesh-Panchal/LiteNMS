package query

import (
	. "datastore/containers"
	. "datastore/utils"
	"go.uber.org/zap"
	"sync"
	"time"
)

func Parser(parserId int, queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, readerRequestChannel chan<- ReaderRequest, readerResponseChannel <-chan map[string]interface{}, parsersWaitGroup *sync.WaitGroup) {

	defer parsersWaitGroup.Done()

	for query := range queryReceiveChannel {

		benchmarkTime := time.Now()

		startDate := query.From - (query.From % 86400)

		endDate := query.To - (query.To % 86400)

		dataType := CounterConfig[query.CounterId][DataType].(string)

		// Total number of days will be: (endDate-startDate)/86400+1
		daysData := make([]map[uint32][]DataPoint, (endDate-startDate)/86400+1)

		requestIndex := 0

		for date := startDate; date <= endDate; date += 86400 {

			request := ReaderRequest{
				ParserId:     parserId,
				RequestIndex: requestIndex,
				StorageKey: StoragePoolKey{
					Date:      UnixToDate(date),
					CounterId: query.CounterId,
				},
				From:      query.From,
				To:        query.To,
				ObjectIds: query.ObjectIds,
			}

			readerRequestChannel <- request

			requestIndex++
		}

		// Listen for response from reader
		for range len(daysData) {

			response := <-readerResponseChannel

			daysData[response["request_index"].(int)] = response["data"].(map[uint32][]DataPoint)

		}

		// Vertical aggregation

		if query.VerticalAggregation != "none" {

			GroupByVerticalAggregator(daysData, query.VerticalAggregation, dataType)

		}

		normalizedDataPoints := make(map[uint32][]DataPoint)

		if query.HorizontalAggregation != "none" {

			normalizedDataPoints = HorizontalAggregator(daysData, query.HorizontalAggregation, dataType, query.Interval)

		} else {

			// Drilldown, Just normalize the days to single slice of datapoints
			for _, day := range daysData {

				for objectId, points := range day {

					normalizedDataPoints[objectId] = append(normalizedDataPoints[objectId], points...)

				}

			}

		}

		Logger.Info("Query result successful in ", zap.Any("ProcessingTime", time.Since(benchmarkTime)), zap.Uint64("queryId", query.QueryId), zap.Any("data", normalizedDataPoints))

		queryResultChannel <- Result{

			query.QueryId,

			normalizedDataPoints,
		}

	}

}
