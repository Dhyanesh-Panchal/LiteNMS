package query

import (
	"context"
	. "datastore/containers"
	. "datastore/storage"
	. "datastore/utils"
	"errors"
	"go.uber.org/zap"
	"sync"
)

type ReaderRequest struct {
	RequestIndex int

	StorageKey StoragePoolKey

	From uint32

	To uint32

	ObjectIds []uint32

	TimeoutContext context.Context
}

func Reader(readerRequestChannel <-chan ReaderRequest, readerResponseChannel chan map[string]interface{}, storagePool *StoragePool, readersWaitGroup *sync.WaitGroup) {

	defer readersWaitGroup.Done()

	for request := range readerRequestChannel {

		storageEngine, err := storagePool.GetStorage(request.StorageKey, false)

		finalDataPoints := make(map[uint32][]DataPoint)

		for _, objectId := range request.ObjectIds {

			finalDataPoints[objectId] = make([]DataPoint, 0)

		}

		if err != nil {

			if errors.Is(err, ErrStorageDoesNotExist) {

				Logger.Info("Storage not present for", zap.Any("storageKey", request.StorageKey))

			}

			// send response with empty data
			readerResponseChannel <- map[string]interface{}{

				"request_index": request.RequestIndex,

				"data": map[uint32][]DataPoint{},

				"error": err,
			}

			continue

		}

		readSingleDay(storageEngine, request.StorageKey, request.ObjectIds, finalDataPoints, request.From, request.To)

		// respond to the QueryParser
		readerResponseChannel <- map[string]interface{}{

			"request_index": request.RequestIndex,

			"data": finalDataPoints,

			"error": nil,
		}

	}

	// channel closed, shutdown is called

	Logger.Info("Reader exiting.")

}

func readSingleDay(storageEngine *Storage, storageKey StoragePoolKey, objectIds []uint32, finalDataPoints map[uint32][]DataPoint, from uint32, to uint32) {

	for _, objectId := range objectIds {

		data, err := storageEngine.Get(objectId)

		if err != nil {

			Logger.Info("Error getting dataPoint ", zap.Uint32("ObjectId", objectId), zap.String("Date", storageKey.Date.Format()), zap.Error(err))

			continue

		}

		dataPoints, err := DeserializeBatch(data, CounterConfig[storageKey.CounterId][DataType].(string))

		if err != nil {

			Logger.Info("Error deserializing dataPoint for objectId: ", zap.Uint32("ObjectId", objectId), zap.String("Date", storageKey.Date.Format()), zap.Error(err))

			continue

		}

		// Append dataPoints if they lie between from and to

		for _, dataPoint := range dataPoints {

			if dataPoint.Timestamp >= from && dataPoint.Timestamp <= to {

				finalDataPoints[objectId] = append(finalDataPoints[objectId], dataPoint)

			}

		}

	}
}
