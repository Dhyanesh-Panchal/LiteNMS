package writer

import (
	. "datastore/containers"
	. "datastore/utils"
	"log"
	"sync"
	"time"
)

var (
	FlushDuration = time.Second * 1
)

func InitWriteHandler(dataWriteChannel <-chan []PolledDataPoint, storagePool *StoragePool, shutdownWaitGroup *sync.WaitGroup) {

	defer shutdownWaitGroup.Done()

	defer log.Println("Write Handler Exiting")

	writersChannel := make(chan WritableObjectBatch, Writers)

	flushRoutineShutdown := make(chan bool)

	var writersWaitGroup sync.WaitGroup

	writersWaitGroup.Add(Writers)

	for i := 0; i < Writers; i++ {

		go writer(writersChannel, storagePool, &writersWaitGroup)

	}

	batchBuffer := NewBatchBuffer()

	go batchBufferFlushRoutine(batchBuffer, writersChannel, flushRoutineShutdown)

	// Listen
	for polledData := range dataWriteChannel {

		for _, dataPoint := range polledData {

			if _, ok := CounterConfig[dataPoint.CounterId]; !ok {

				// Invalid counterId, skip
				log.Println("Bad CounterID, Dropping dataPoint:", dataPoint)

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

	// Channel Closed, Shutting down writer

	flushRoutineShutdown <- true

	// Wait for final flush
	<-flushRoutineShutdown

	// Close writers
	close(writersChannel)

	writersWaitGroup.Wait()

}

func batchBufferFlushRoutine(batchBuffer *BatchBuffer, writersChannel chan<- WritableObjectBatch, flushRoutineShutdown chan bool) {

	for {

		select {

		case <-flushRoutineShutdown:

			// Flush present entries and exit

			if !batchBuffer.EmptyBuffer {

				batchBuffer.Flush(writersChannel)

			}

			batchBuffer.flushTicker.Stop()

			flushRoutineShutdown <- true

			return

		case <-batchBuffer.flushTicker.C:

			if !batchBuffer.EmptyBuffer {

				batchBuffer.Flush(writersChannel)

			}
		}
	}
}
