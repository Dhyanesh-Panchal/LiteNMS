package query

import (
	. "datastore/containers"
	. "datastore/utils"
	"sync"
)

type Query struct {
	QueryId uint64 `json:"query_id" msgpack:"query_id"`

	From uint32 `json:"from" msgpack:"from"`

	To uint32 `json:"to" msgpack:"to"`

	ObjectIds []uint32 `json:"object_ids" msgpack:"object_ids"`

	CounterId uint16 `json:"counter_id" msgpack:"counter_id"`

	ObjectWiseAggregation string `json:"object_wise_aggregation" msgpack:"object_wise_aggregation"`

	TimestampAggregation string `json:"timestamp_aggregation" msgpack:"timestamp_aggregation"`

	Interval uint32 `json:"interval" msgpack:"interval"`
}

type Result struct {
	QueryId uint64 `json:"query_id" msgpack:"query_id"`

	Data map[uint32][]DataPoint `json:"data" msgpack:"data"`

	Error string `json:"error" msgpack:"error"`
}

func InitQueryEngine(queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, storagePool *StoragePool, shutdownWaitGroup *sync.WaitGroup) {

	defer shutdownWaitGroup.Done()

	// Spawn Query Parsers

	var parsersWaitGroup sync.WaitGroup

	parsersWaitGroup.Add(QueryParsers)

	for range QueryParsers {

		go Parser(queryReceiveChannel, queryResultChannel, storagePool, &parsersWaitGroup)

	}

	parsersWaitGroup.Wait()

	close(queryResultChannel)

}
