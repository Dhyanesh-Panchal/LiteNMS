package query

import (
	. "datastore/containers"
	. "datastore/storage"
	. "datastore/utils"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"sync"
)

type ReaderRequest struct {
	ParserId int

	RequestIndex int

	StorageKey StoragePoolKey

	From uint32

	To uint32

	ObjectIds []uint32
}

func Reader(readerRequestChannel <-chan ReaderRequest, parserWaitChannels []chan map[string]interface{}, storagePool *StoragePool, readersWaitGroup *sync.WaitGroup) {

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

			// send response
			parserWaitChannels[request.ParserId] <- map[string]interface{}{

				"requestIndex": request.RequestIndex,

				"data": map[uint32][]DataPoint{},

				"error": err,
			}

			continue

		}

		readSingleDay(storageEngine, request.StorageKey, request.ObjectIds, finalDataPoints, request.From, request.To)

		fmt.Println("Data for ", request.ParserId, request.RequestIndex, request.StorageKey, request.ObjectIds)

		// respond to the Parser
		parserWaitChannels[request.ParserId] <- map[string]interface{}{

			"requestIndex": request.RequestIndex,

			"data": finalDataPoints,

			"error": nil,
		}

	}

	// channel closed, shutdown is called

	Logger.Info("Reader exiting.")

}

//func queryHistogram(from uint32, to uint32, counterId uint16, objects []uint32, storagePool *StoragePool) (map[uint32][]DataPoint, error) {
//
//	finalData := map[uint32][]DataPoint{}
//
//	for date := from - (from % 86400); date <= to; date += 86400 {
//
//		dateObject := UnixToDate(date)
//
//		storageKey := StoragePoolKey{
//
//			Date: dateObject,
//
//			CounterId: counterId,
//		}
//
//		storageEngine, err := storagePool.GetStorage(storageKey, false)
//
//		if err != nil {
//
//			if errors.Is(err, ErrStorageDoesNotExist) {
//
//				Logger.Info("Storage not present for date:"+dateObject.Format(), zap.Uint16("counterID:", counterId))
//
//				continue
//
//			}
//
//			return nil, err
//
//		}
//
//		readSingleDay(dateObject, storageEngine, counterId, objects, finalData, from, to)
//
//		if storageKey.Date.Day != time.Now().Day() {
//
//			// Close storage for days other than current day.
//			// current day's storage is constantly used by writer hence no need to close it.
//
//			storagePool.CloseStorage(storageKey)
//		}
//
//	}
//
//	return finalData, nil
//
//}

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
