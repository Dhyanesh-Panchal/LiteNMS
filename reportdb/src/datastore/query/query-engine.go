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

	Error string `json:"error"`
}

func InitQueryEngine(queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, storagePool *StoragePool, shutdownWaitGroup *sync.WaitGroup) {

	defer shutdownWaitGroup.Done()

	// Spawn Query Parsers

	var parsersWaitGroup sync.WaitGroup

	parsersWaitGroup.Add(QueryParsers)

	for range QueryParsers {

		go QueryParser(queryReceiveChannel, queryResultChannel, storagePool, &parsersWaitGroup)

	}

	parsersWaitGroup.Wait()

	close(queryResultChannel)

}
