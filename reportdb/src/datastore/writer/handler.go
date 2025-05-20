package writer

import (
	. "datastore/containers"
	. "datastore/utils"
	"go.uber.org/zap"
	"sync"
	"time"
)

var (
	FlushDuration = time.Second * 5
)

func InitWriteHandler(dataWriteChannel <-chan []PolledDataPoint, storagePool *StoragePool, shutdownWaitGroup *sync.WaitGroup) {

	defer shutdownWaitGroup.Done()

	defer Logger.Info("Write Handler Exiting")

	writersChannel := make(chan WritableObjectBatch, Writers)

	flushRoutineShutdown := make(chan bool)

	var writersWaitGroup sync.WaitGroup

	writersWaitGroup.Add(Writers)

	for range Writers {

		go writer(writersChannel, storagePool, &writersWaitGroup)

	}

	batchBuffer := NewBatchBuffer()

	go batchBufferFlushRoutine(batchBuffer, writersChannel, flushRoutineShutdown)

	// Listen
	for polledData := range dataWriteChannel {

		for _, dataPoint := range polledData {

			if _, ok := CounterConfig[dataPoint.CounterId]; !ok {

				// Invalid counterId, skip
				Logger.Info("bad counterId, dropping dataPoint.", zap.Uint16("counterId", dataPoint.CounterId), zap.Any("dataPoint", dataPoint))

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
