package writer

import (
	. "datastore/containers"
	. "datastore/utils"
	"go.uber.org/zap"
	"sync"
)

type WritableObjectBatch struct {
	StorageKey StoragePoolKey
	ObjectId   uint32
	Values     []DataPoint
}

func writer(writersChannel <-chan WritableObjectBatch, storagePool *StoragePool, writerWaitGroup *sync.WaitGroup) {

	defer writerWaitGroup.Done()

	dataBytesContainer := make([]byte, 0)

	for dataBatch := range writersChannel {

		//Logger.Debug("writer received data", zap.Any("dataBatch", dataBatch))

		// Serialize the Data

		if err := SerializeBatch(dataBatch.Values, &dataBytesContainer, CounterConfig[dataBatch.StorageKey.CounterId][DataType]); err != nil {

			Logger.Error("error serializing the batch", zap.Error(err))

		}

		storageEngine, err := storagePool.GetStorage(dataBatch.StorageKey, true)

		if err != nil {

			Logger.Error("error acquiring storage engine for writing", zap.Error(err))

		}

		err = storageEngine.Put(dataBatch.ObjectId, dataBytesContainer)

		if err != nil {

			Logger.Error("error writing to storage:", zap.Error(err))

		}

		// Clear the cache for this object
		DataPointsCache.Del(CreateCacheKey(dataBatch.StorageKey, dataBatch.ObjectId))

		// reslice the dataBytesContainer
		dataBytesContainer = dataBytesContainer[:0]

	}

	Logger.Info("Writer exiting.")
}
