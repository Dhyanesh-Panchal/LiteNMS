package writer

import (
	. "datastore/containers"
	"sync"
	"time"
)

type BatchBuffer struct {
	buffer map[StoragePoolKey]map[uint32][]DataPoint // StoragePoolKey -> {Date,CounterId},

	flushTicker *time.Ticker

	EmptyBuffer bool

	flushLock sync.RWMutex
}

func NewBatchBuffer() *BatchBuffer {

	pool := make(map[StoragePoolKey]map[uint32][]DataPoint)

	flushTicker := time.NewTicker(FlushDuration)

	return &BatchBuffer{

		buffer: pool,

		flushTicker: flushTicker,

		EmptyBuffer: true,
	}

}

func (buffer *BatchBuffer) AddDataPoint(key StoragePoolKey, objectId uint32, dataPoint DataPoint) {

	buffer.flushLock.Lock()

	defer buffer.flushLock.Unlock()

	buffer.EmptyBuffer = false

	if _, ok := buffer.buffer[key]; !ok {

		buffer.buffer[key] = make(map[uint32][]DataPoint)

	}

	buffer.buffer[key][objectId] = append(buffer.buffer[key][objectId], dataPoint)

}

func (buffer *BatchBuffer) GetDataPoints(key StoragePoolKey, objectId uint32) []DataPoint {

	buffer.flushLock.RLock()

	defer buffer.flushLock.RUnlock()

	return buffer.buffer[key][objectId]
}

func (buffer *BatchBuffer) Flush(dataChannel chan<- WritableObjectBatch) {

	buffer.flushLock.Lock()

	defer buffer.flushLock.Unlock()

	for storageKey, objects := range buffer.buffer {

		for objectId, dataPoints := range objects {

			objectData := WritableObjectBatch{
				storageKey,

				objectId,

				dataPoints,
			}

			dataChannel <- objectData

			// Empty the buffer for that object
			delete(objects, objectId)

		}

	}

	buffer.EmptyBuffer = true

}

func batchBufferFlushRoutine(batchBuffer *BatchBuffer, writersChannel chan<- WritableObjectBatch, flushRoutineShutdown chan struct{}) {

	for {

		select {

		case <-flushRoutineShutdown:

			// Flush present entries and exit

			if !batchBuffer.EmptyBuffer {

				batchBuffer.Flush(writersChannel)

			}

			batchBuffer.flushTicker.Stop()

			flushRoutineShutdown <- struct{}{}

			return

		case <-batchBuffer.flushTicker.C:

			if !batchBuffer.EmptyBuffer {

				batchBuffer.Flush(writersChannel)

			}
		}
	}
}
