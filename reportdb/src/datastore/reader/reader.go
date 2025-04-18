package reader

import (
	. "datastore/containers"
	. "datastore/storage"
	. "datastore/utils"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"log"
	"sync"
	"time"
)

func reader(queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, storagePool *StoragePool, readersWaitGroup *sync.WaitGroup) {

	defer readersWaitGroup.Done()

	for query := range queryReceiveChannel {

		startTime := time.Now()

		result, err := queryHistogram(query.From, query.To, query.CounterId, query.ObjectIds, storagePool)

		if err != nil {

			log.Printf("Error querying datastore: %s" + err.Error())

		}

		queryResultChannel <- Result{

			QueryId: query.QueryId,

			Data: result,
		}

		dataPoints := 0

		for _, resultPoint := range result {
			dataPoints += len(resultPoint)
		}

		fmt.Println("Total Data Points: ", dataPoints, "In:", time.Since(startTime))

	}

	// channel closed, shutdown is called

	Logger.Info("Reader exiting.")

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

				Logger.Info("Storage not present for date:"+dateObject.Format(), zap.Uint16("counterID:", counterId))

				continue

			}

			return nil, err

		}

		readSingleDay(dateObject, storageEngine, counterId, objects, finalData, from, to)

		if storageKey.Date.Day != time.Now().Day() {

			// Close storage for days other than current day.
			// current day's storage is constantly used by writer hence no need to close it.

			storagePool.CloseStorage(storageKey)
		}

	}

	return finalData, nil

}

// readSingleDay Note: readFullDate function changes the state of the finalData; hence if run in parallel, proper synchronization is needed.
func readSingleDay(date Date, storageEngine *Storage, counterId uint16, objects []uint32, finalData map[uint32][]DataPoint, from uint32, to uint32) {

	for _, objectId := range objects {

		data, err := storageEngine.Get(objectId)

		if err != nil {

			Logger.Info("Error getting dataPoint ", zap.Uint32("ObjectId", objectId), zap.String("Date", date.Format()))
			continue

		}

		dataPoints, err := DeserializeBatch(data, CounterConfig[counterId][DataType].(string))

		if err != nil {

			Logger.Info("Error deserializing dataPoint for objectId: ", zap.Uint32("ObjectId", objectId), zap.String("Date", date.Format()), zap.Error(err))

			continue

		}

		// Append dataPoints if they lie between from and to

		for _, dataPoint := range dataPoints {

			if dataPoint.Timestamp >= from && dataPoint.Timestamp <= to {

				finalData[objectId] = append(finalData[objectId], dataPoint)

			}

		}

	}
}
