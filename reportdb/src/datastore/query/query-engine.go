package query

import (
	. "datastore/containers"
	. "datastore/utils"
	"sync"
)

type Query struct {
	QueryId               uint64   `json:"query_id"`
	From                  uint32   `json:"from"`
	To                    uint32   `json:"to"`
	ObjectIds             []uint32 `json:"object_ids"`
	CounterId             uint16   `json:"counter_id"`
	VerticalAggregation   string   `json:"vertical_aggregation"`
	HorizontalAggregation string   `json:"horizontal_aggregation"`
	Interval              uint32   `json:"interval"`
}

type Result struct {
	QueryId uint64 `json:"query_id"`

	Data interface{} `json:"data"`
}

func InitQueryEngine(queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, storagePool *StoragePool, shutdownWaitGroup *sync.WaitGroup) {

	defer shutdownWaitGroup.Done()

	// Structures for communication between parsers and readers

	readerRequestChannel := make(chan ReaderRequest, 100) // TODO: Shift channel size to config

	parserWaitChannels := make([]chan map[string]interface{}, QueryParsers)

	for parserId := range QueryParsers {

		parserWaitChannels[parserId] = make(chan map[string]interface{}, 10) // TODO: Shift channel size to config

	}

	// Spawn Readers

	var readersWaitGroup sync.WaitGroup

	readersWaitGroup.Add(Readers)

	for range Readers {

		go Reader(readerRequestChannel, parserWaitChannels, storagePool, &readersWaitGroup)

	}

	// Spawn Query Parsers

	var parsersWaitGroup sync.WaitGroup

	parsersWaitGroup.Add(QueryParsers)

	for parserId := range QueryParsers {

		go Parser(parserId, queryReceiveChannel, queryResultChannel, readerRequestChannel, parserWaitChannels[parserId], &parsersWaitGroup)

	}

	parsersWaitGroup.Wait()

	close(readerRequestChannel)

	readersWaitGroup.Wait()

	close(queryResultChannel)

}
