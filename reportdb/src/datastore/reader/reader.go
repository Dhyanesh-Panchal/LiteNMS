package reader

import (
	. "datastore/containers"
	. "datastore/storage"
	. "datastore/utils"
	"errors"
	"log"
	"sync"
)

func InitQueryHandler(queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, storagePool *StoragePool, shutdownWaitGroup *sync.WaitGroup) {

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

				continue

			}

			return nil, err

		}

		readSingleDay(dateObject, storageEngine, counterId, objects, finalData, from, to)

		storagePool.CloseStorage(storageKey)

	}

	return finalData, nil

}

// readSingleDay Note: readFullDate function changes the state of the finalData; hence if run in parallel, proper synchronization is needed.
func readSingleDay(date Date, storageEngine *Storage, counterId uint16, objects []uint32, finalData map[uint32][]DataPoint, from uint32, to uint32) {

	for _, objectId := range objects {

		data, err := storageEngine.Get(objectId)

		if err != nil {

			log.Println("Error getting data for objectId: ", objectId, " Day: ", date)
			continue

		}

		dataPoints, err := DeserializeBatch(data, CounterConfig[counterId][DataType].(string))

		if err != nil {

			log.Println("Error deserializing data for objectId: ", objectId, "Day: ", date, "Error:", err)

			continue

		}

		// Append dataPoints if they lie between from and to

		for _, data := range dataPoints {

			if data.Timestamp >= from && data.Timestamp <= to {

				finalData[objectId] = append(finalData[objectId], data)

			}

		}

	}
}
