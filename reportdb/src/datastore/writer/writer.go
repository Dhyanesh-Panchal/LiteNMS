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

	for dataBatch := range writersChannel {

		Logger.Info("writer received data", zap.Any("dataBatch", dataBatch))

		// Serialize the Data
		data, err := SerializeBatch(dataBatch.Values, CounterConfig[dataBatch.StorageKey.CounterId][DataType].(string))

		if err != nil {

			Logger.Error("error serializing the batch", zap.Error(err))

		}

		storageEngine, err := storagePool.GetStorage(dataBatch.StorageKey, true)

		if err != nil {

			Logger.Error("error acquiring storage engine for writing", zap.Error(err))

		}

		err = storageEngine.Put(dataBatch.ObjectId, data)

		if err != nil {

			Logger.Error("error writing to storage:", zap.Error(err))

		}
	}

	Logger.Info("Writer exiting.")
}
