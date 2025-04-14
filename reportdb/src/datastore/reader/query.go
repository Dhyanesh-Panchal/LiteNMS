package reader

import (
	. "datastore/containers"
	. "datastore/storage"
	"errors"
	"log"
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

	for query := range queryReceiveChannel {

		result, err := queryHistogram(query.From, query.To, query.CounterId, query.ObjectIds, storagePool)

		if err != nil {

			log.Printf("Error querying datastore: %s", err)

		}

		queryResultChannel <- Result{

			QueryId: query.QueryId,

			Data: result,
		}

	}

	close(queryResultChannel)

}

func queryHistogram(from uint32, to uint32, counterId uint16, objects []uint32, storagePool *StoragePool) (map[uint32][]DataPoint, error) {

	finalData := map[uint32][]DataPoint{}

	for date := from - (from % 86400); date <= to; date += 86400 {

		dateObject := UnixToDate(date)

		storageKey := StoragePoolKey{

			Date: dateObject,

			CounterId: counterId,
		}

		storageEngine, err := storagePool.GetStorage(storageKey, false)

		if err != nil {

			if errors.Is(err, ErrStorageDoesNotExist) {

				log.Println("Storage not present for date:", dateObject)

				continue

			}

			return nil, err

		}

		readSingleDay(dateObject, storageEngine, counterId, objects, finalData, from, to)

		storagePool.CloseStorage(storageKey)

	}

	return finalData, nil

}
