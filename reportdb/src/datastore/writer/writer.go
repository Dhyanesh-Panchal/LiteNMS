package writer

import (
	. "datastore/containers"
	. "datastore/utils"
	"log"
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

		log.Println(dataBatch.Values)

		// Serialize the Data
		data, err := SerializeBatch(dataBatch.Values, CounterConfig[dataBatch.StorageKey.CounterId][DataType].(string))

		if err != nil {

			log.Println("Error serializing the batch", err)

		}

		storageEngine, err := storagePool.GetStorage(dataBatch.StorageKey, true)

		if err != nil {

			log.Println("Error acquiring storage engine for writing", err)

		}

		err = storageEngine.Put(dataBatch.ObjectId, data)

		if err != nil {

			log.Println("Error writing to storage:", err)

		}
	}

	log.Println("Writer exiting.")
}
