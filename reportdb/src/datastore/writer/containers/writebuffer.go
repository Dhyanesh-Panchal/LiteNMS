package containers

import (
	. "datastore/containers"
	"sync"
)

type WriteBuffer struct {
	buffer map[StoragePoolKey]map[uint32][]DataPoint

	flushLock sync.RWMutex

	EmptyBuffer bool
}

func NewWritePool() *WriteBuffer {

	pool := make(map[StoragePoolKey]map[uint32][]DataPoint)

	return &WriteBuffer{

		buffer: pool,

		EmptyBuffer: true,
	}

}

func (buffer *WriteBuffer) AddDataPoint(key StoragePoolKey, objectId uint32, dataPoint DataPoint) {

	buffer.flushLock.Lock()
	defer buffer.flushLock.Unlock()

	buffer.EmptyBuffer = false

	if _, ok := buffer.buffer[key]; !ok {

		buffer.buffer[key] = make(map[uint32][]DataPoint)

	}

	buffer.buffer[key][objectId] = append(buffer.buffer[key][objectId], dataPoint)

}

func (buffer *WriteBuffer) GetDataPoints(key StoragePoolKey, objectId uint32) []DataPoint {

	buffer.flushLock.RLock()

	defer buffer.flushLock.RUnlock()

	return buffer.buffer[key][objectId]
}

func (buffer *WriteBuffer) Flush(dataChannel chan<- WritableObjectData) {

	buffer.flushLock.Lock()

	defer buffer.flushLock.Unlock()

	for storageKey, objects := range buffer.buffer {

		for objectId, dataPoints := range objects {

			objectData := WritableObjectData{
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
