package reader

import (
	. "datastore/containers"
	. "datastore/utils"
	"sync"
)

type Query struct {
	QueryId     uint64   `json:"query_id"`
	From        uint32   `json:"from"`
	To          uint32   `json:"to"`
	ObjectIds   []uint32 `json:"object_ids"`
	CounterId   uint16   `json:"counter_id"`
	Aggregation string   `json:"aggregation"`
}

type Result struct {
	QueryId uint64 `json:"query_id"`

	Data map[uint32][]DataPoint `json:"data"`
}

func InitQueryEngine(queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, storagePool *StoragePool, shutdownWaitGroup *sync.WaitGroup) {

	defer shutdownWaitGroup.Done()

	var readersWaitGroup sync.WaitGroup

	readersWaitGroup.Add(Readers)

	for range Readers {

		go reader(queryReceiveChannel, queryResultChannel, storagePool, &readersWaitGroup)

	}

	//for query := range queryReceiveChannel {
	//
	//	result, err := queryHistogram(query.From, query.To, query.CounterId, query.ObjectIds, storagePool)
	//
	//	if err != nil {
	//
	//		log.Printf("Error querying datastore: %s", err)
	//
	//	}
	//
	//	queryResultChannel <- Result{
	//
	//		QueryId: query.QueryId,
	//
	//		Data: result,
	//	}
	//
	//}

	readersWaitGroup.Wait()

	close(queryResultChannel)

}
