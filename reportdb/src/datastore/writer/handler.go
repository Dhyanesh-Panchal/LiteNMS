package writer

import (
	. "datastore/containers"
	. "datastore/utils"
	"log"
	"time"
)

var (
	FlushDuration = time.Second * 1
)

func InitWriteHandler(dataWriteChannel <-chan []PolledDataPoint, storagePool *StoragePool) {

	writersChannel := make(chan WritableObjectBatch, Writers)

	writersShutdown := make(chan bool, Writers)

	flushRoutineShutdown := make(chan bool)

	for i := 0; i < Writers; i++ {

		go writer(writersChannel, storagePool, writersShutdown)

	}

	batchBuffer := NewBatchBuffer()

	go batchBufferFlushRoutine(batchBuffer, writersChannel, flushRoutineShutdown)

	// Listen
	for {
		select {
		case <-GlobalShutdown:

			flushRoutineShutdown <- true

			// Wait for final flush
			<-flushRoutineShutdown

			for range Writers {

				writersShutdown <- true

			}

			return

		case polledData := <-dataWriteChannel:

			for _, dataPoint := range polledData {

				if _, ok := CounterConfig[dataPoint.CounterId]; !ok {

					// Invalid counterId, skip
					log.Println("Bad CounterID for dataPoint:", dataPoint)
					
					continue

				}

				storageKey := StoragePoolKey{

					Date: UnixToDate(dataPoint.Timestamp),

					CounterId: dataPoint.CounterId,
				}

				batchBuffer.AddDataPoint(

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

func batchBufferFlushRoutine(batchBuffer *BatchBuffer, writersChannel chan<- WritableObjectBatch, flushRoutineShutdown chan bool) {

	flushTicker := time.NewTicker(FlushDuration)

	for {

		select {

		case <-flushRoutineShutdown:

			// Flush present entries and exit

			if !batchBuffer.EmptyBuffer {

				batchBuffer.Flush(writersChannel)

			}

			flushTicker.Stop()

			flushRoutineShutdown <- true

			return

		case <-flushTicker.C:

			if !batchBuffer.EmptyBuffer {

				batchBuffer.Flush(writersChannel)

			}
		}
	}
}
