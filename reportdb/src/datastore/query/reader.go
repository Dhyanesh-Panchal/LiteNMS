package query

import (
	"context"
	. "datastore/containers"
	. "datastore/storage"
	. "datastore/utils"
	"errors"
	"go.uber.org/zap"
	"sync"
	"time"
)

type ReaderRequest struct {
	RequestIndex int

	StorageKey StoragePoolKey

	From uint32

	To uint32

	ObjectIds []uint32

	TimeoutContext context.Context
}

type ReaderResponse struct {
	RequestIndex int

	Data map[uint32][]DataPoint

	Error error
}

func Reader(readerRequestChannel <-chan ReaderRequest, readerResponseChannel chan ReaderResponse, storagePool *StoragePool, readersWaitGroup *sync.WaitGroup) {

	defer readersWaitGroup.Done()

	for request := range readerRequestChannel {

		storageEngine, err := storagePool.GetStorage(request.StorageKey, false)

		if err != nil {

			if errors.Is(err, ErrStorageDoesNotExist) {

				Logger.Info("Storage not present for", zap.Any("storageKey", request.StorageKey))

			}

			// send response with empty data
			readerResponseChannel <- ReaderResponse{

				request.RequestIndex,

				nil,

				err,
			}

			continue

		}

		data, err := readSingleDay(storageEngine, request.StorageKey, request.ObjectIds, request.From, request.To)

		if err != nil {

			readerResponseChannel <- ReaderResponse{

				request.RequestIndex,

				nil,

				err,
			}

		} else {

			// respond to the Parser with day's data

			readerResponseChannel <- ReaderResponse{

				request.RequestIndex,

				data,

				nil,
			}

		}

	}

	// channel closed, shutdown is called

	Logger.Info("Reader exiting.")

}

func readSingleDay(storageEngine *Storage, storageKey StoragePoolKey, objectIds []uint32, from uint32, to uint32) (map[uint32][]DataPoint, error) {

	if len(objectIds) == 0 {

		// No objectIds mentioned, hence get all objectIds

		var err error

		objectIds, err = storageEngine.GetAllKeys()

		if err != nil {

			Logger.Error("Error getting all storage keys", zap.Error(err))

			return nil, err

		}

	}

	finalDataPoints := make(map[uint32][]DataPoint)

	for _, objectId := range objectIds {

		var dataPoints []DataPoint

		data, hit := DataPointsCache.Get(CreateCacheKey(storageKey, objectId))

		if !hit {

			data, err := storageEngine.Get(objectId)

			if err != nil {

				Logger.Info("Error getting dataPoint ", zap.Uint32("ObjectId", objectId), zap.String("Date", storageKey.Date.Format()), zap.Error(err))

				continue

			}

			dataPoints, err = DeserializeBatch(data, CounterConfig[storageKey.CounterId][DataType].(string))

			if err != nil {

				Logger.Info("Error deserializing dataPoint for objectId: ", zap.Uint32("ObjectId", objectId), zap.String("Date", storageKey.Date.Format()), zap.Error(err))

				continue

			}

			// Don't set cache for current day
			if UnixToDate(time.Now().Unix()) != storageKey.Date {

				DataPointsCache.Set(CreateCacheKey(storageKey, objectId), dataPoints, 0)

			}

		} else {

			//Logger.Debug("Cache hit for:", zap.Uint32("ObjectId", objectId), zap.String("Date", storageKey.Date.Format()))

			dataPoints = data.([]DataPoint)

		}

		// Append dataPoints if they lie between from and to

		for _, dataPoint := range dataPoints {

			if dataPoint.Timestamp >= from && dataPoint.Timestamp <= to {

				finalDataPoints[objectId] = append(finalDataPoints[objectId], dataPoint)

			}

		}

	}

	return finalDataPoints, nil
}
