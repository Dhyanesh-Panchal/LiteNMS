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

			// respond to the QueryParser with day's data

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

	finalDataPoints := make(map[uint32][]DataPoint)

	if len(objectIds) == 0 {

		// No objectIds mentioned, hence get data for all

		data, err := storageEngine.GetAll()

		if err != nil {

			Logger.Info("Error getting data for all objects for", zap.Any("storageKey", storageKey), zap.Error(err))

			return nil, err

		}

		for objectId, objectData := range data {

			dataPoints, err := DeserializeBatch(objectData, CounterConfig[storageKey.CounterId][DataType].(string))

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

		return finalDataPoints, nil

	} else {

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

		return finalDataPoints, nil

	}

}
