package writer

import (
	"log"
	. "reportdb/config"
	. "reportdb/containers"
	. "reportdb/global"
	. "reportdb/writer/containers"
	"time"
)

func InitWriter(dataChannel <-chan []PolledDataPoint, storagePool *StoragePool) {

	writersChannel := make(chan WritableObjectData, WriterCount)

	for i := 0; i < WriterCount; i++ {

		go writeWorker(writersChannel, storagePool)

	}

	writeBuffer := NewWritePool()

	go WriteBufferFlushRoutine(writeBuffer, writersChannel)

	// Listen
	for {
		select {
		case <-GlobalShutdown:

		case polledData := <-dataChannel:

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

}

func writeWorker(writersChannel <-chan WritableObjectData, storagePool *StoragePool) {
	for {
		select {
		case dataBatch := <-writersChannel:

			// Serialize the Data

			data := SerializeBatch(dataBatch.Values)

			storageEngine, err := storagePool.AcquireStorage(dataBatch.StorageKey)

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

	flushTicker := time.NewTicker(WriterFlushDuration)

	for {
		select {

		case <-flushTicker.C:

			if !dataWriteBuffer.EmptyBuffer {

				dataWriteBuffer.Flush(writersChannel)

			}
		}
	}
}
