package writer

import (
	. "datastore/containers"
	. "datastore/utils"
	. "datastore/writer/containers"
	"log"
	"time"
)

var (
	FlushDuration = time.Second * 1
)

func InitWriter(dataWriteChannel <-chan []PolledDataPoint, storagePool *StoragePool) {

	writersChannel := make(chan WritableObjectData, Writers)
	writersShutdown := make(chan bool, Writers)

	for i := 0; i < Writers; i++ {

		go writeWorker(writersChannel, storagePool)

	}

	writeBuffer := NewWritePool()

	go WriteBufferFlushRoutine(writeBuffer, writersChannel)

	// Listen
	for {
		select {
		case <-GlobalShutdown:
			break

		case polledData := <-dataWriteChannel:

			for _, dataPoint := range polledData {

				storageKey := StoragePoolKey{
					Date:      UnixToDate(dataPoint.Timestamp),
					CounterId: dataPoint.CounterId,
				}

				writeBuffer.AddDataPoint(

					storageKey,

					dataPoint.ObjectId,

					DataPoint{

						Timestamp: dataPoint.Timestamp,

						Value: dataPoint.Value,
					},
				)

			}

		}

	}

	// Global Shutdown sequence

}

func writeWorker(writersChannel <-chan WritableObjectData, storagePool *StoragePool, shutdownChannel chan bool) {
	for {
		select {
		case <-shutdownChannel:

		case dataBatch := <-writersChannel:

			// Serialize the Data
			data, err := SerializeBatch(dataBatch.Values, CounterConfig[dataBatch.StorageKey.CounterId][DataType].(string))

			storageEngine, err := storagePool.GetStorage(dataBatch.StorageKey, true)

			if err != nil {

				log.Println("Error acquiring storage engine for writing", err)

			}

			err = storageEngine.Put(dataBatch.ObjectId, data)

			if err != nil {

				log.Println("Error writing to storage:", err)

			}
		}
	}
}

func WriteBufferFlushRoutine(dataWriteBuffer *WriteBuffer, writersChannel chan<- WritableObjectData) {

	flushTicker := time.NewTicker(FlushDuration)

	for {
		select {
		case <-GlobalShutdown:

			// Flush present entries and exit

			if !dataWriteBuffer.EmptyBuffer {

				dataWriteBuffer.Flush(writersChannel)

			}

			flushTicker.Stop()

			break

		case <-flushTicker.C:

			if !dataWriteBuffer.EmptyBuffer {

				dataWriteBuffer.Flush(writersChannel)

			}
		}
	}
}
