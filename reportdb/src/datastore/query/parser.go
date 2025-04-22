package query

import (
	. "datastore/containers"
)

func Parser(parserId int, queryReceiveChannel <-chan map[string]interface{}, queryResultChannel chan<- Result, readerRequestChannel chan<- ReaderRequest, readerResponseChannel <-chan map[string]interface{}) {

	for query := range queryReceiveChannel {

		from := uint32(query["from"].(float64))

		to := uint32(query["to"].(float64))

		startDate := from - (from % 86400)

		endDate := to - (to % 86400)

		counterId := uint16(query["counterId"].(float64))

		objectIds := query["objectIds"].([]uint32)

		// Total number of days will be: (endDate-startDate)/86400+1
		daysData := make([]map[uint32][]DataPoint, (endDate-startDate)/86400+1)

		requestIndex := 0

		for date := startDate; date <= endDate; date += 86400 {

			request := ReaderRequest{
				ParserId:     parserId,
				RequestIndex: requestIndex,
				StorageKey: StoragePoolKey{
					Date:      UnixToDate(date),
					CounterId: counterId,
				},
				From:      from,
				To:        to,
				ObjectIds: objectIds,
			}

			readerRequestChannel <- request

			requestIndex++
		}

		// Listen for response
		for range len(daysData) {

			response := <-readerResponseChannel

			daysData[response["requestIndex"].(int)] = response["data"].(map[uint32][]DataPoint)

		}

		// merge all the days in single structure
		//mergedDatapoints := make(map[uint32][]DataPoint)
		//
		//for _, day := range daysData {
		//
		//}

		queryResultChannel <- Result{
			uint64(query["queryId"].(float64)),
			daysData,
		}

	}

}

func parseQuery(query map[string]interface{}) {

	// recognise the days based on from-to

}
