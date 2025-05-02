package query

import (
	"context"
	. "datastore/containers"
	. "datastore/utils"
	"go.uber.org/zap"
	"sync"
	"time"
)

func QueryParser(queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, storagePool *StoragePool, parsersWaitGroup *sync.WaitGroup) {

	defer parsersWaitGroup.Done()

	// Initialize Readers

	readerRequestChannel := make(chan ReaderRequest, 10) // TODO: Shift channel size to config

	readerResponseChannel := make(chan ReaderResponse, 10)

	var readersWaitGroup sync.WaitGroup

	readersWaitGroup.Add(Readers)

	for range Readers {

		go Reader(readerRequestChannel, readerResponseChannel, storagePool, &readersWaitGroup)

	}

	// Listen for query

	for query := range queryReceiveChannel {

		Logger.Info("Query received: ", zap.Any("query", query))

		benchmarkTime := time.Now()

		queryTimeoutContext, queryTimeoutContextCancel := context.WithTimeout(context.Background(), time.Duration(QueryTimeoutTime)*time.Second)

		startDate := query.From - (query.From % 86400)

		endDate := query.To - (query.To % 86400)

		dataType := CounterConfig[query.CounterId][DataType].(string)

		// Total number of days will be: (endDate-startDate)/86400+1
		daysData := make([]map[uint32][]DataPoint, (endDate-startDate)/86400+1)

		requestIndex := 0

		for date := startDate; date <= endDate; date += 86400 {

			select {

			case <-queryTimeoutContext.Done():

				break

			case readerRequestChannel <- ReaderRequest{

				RequestIndex: requestIndex,

				StorageKey: StoragePoolKey{
					Date:      UnixToDate(date),
					CounterId: query.CounterId,
				},

				From: query.From,

				To: query.To,

				ObjectIds: query.ObjectIds,

				TimeoutContext: queryTimeoutContext,
			}:

				requestIndex++
			}

		}

		// Listen for response from reader
		for range len(daysData) {

			select {

			case <-queryTimeoutContext.Done():

				break

			case response := <-readerResponseChannel:

				if response.Error == nil {

					daysData[response.RequestIndex] = response.Data

				}

			}

		}

		// If the datatype is string, there is no point of aggregation. Hence for string queries, just normalize the days and send the drilldown.

		// Vertical aggregation

		if query.VerticalAggregation != "none" && dataType != "string" {

			GroupByVerticalAggregator(daysData, query.VerticalAggregation, queryTimeoutContext)

		}

		// Necessary structures initialization

		normalizedDataPoints := make(map[uint32][]DataPoint)

		if query.HorizontalAggregation != "none" && dataType != "string" {

			HorizontalAggregator(daysData, query.HorizontalAggregation, query.Interval, query.From, normalizedDataPoints, queryTimeoutContext)

		} else {

			// Drilldown, Just normalize the days to single slice of dataPoints
			for _, day := range daysData {

				select {

				case <-queryTimeoutContext.Done():

					break

				default:

					for objectId, points := range day {

						normalizedDataPoints[objectId] = append(normalizedDataPoints[objectId], points...)

					}

				}

			}

		}

		select {
		case <-queryTimeoutContext.Done():

			Logger.Info("Query timed out.", zap.Uint64("queryId", query.QueryId))

			queryResultChannel <- Result{

				query.QueryId,

				nil,

				"query timed out",
			}

		default:

			Logger.Info("Query result successful in ", zap.Any("ProcessingTime", time.Since(benchmarkTime)), zap.Uint64("queryId", query.QueryId), zap.Any("data-points", normalizedDataPoints))

			queryResultChannel <- Result{

				query.QueryId,

				normalizedDataPoints,

				"",
			}

		}

		queryTimeoutContextCancel()

	}

	close(readerRequestChannel)

	readersWaitGroup.Wait()

}
