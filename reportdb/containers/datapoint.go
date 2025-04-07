package containers

import (
	"encoding/binary"
	"math"
)

type DataPoint struct {
	Timestamp uint32
	Value     interface{}
}

func SerializeBatch(data []DataPoint) []byte {

	if len(data) == 0 {

		return nil

	}

	if _, isString := data[0].Value.(string); isString {

		return serializeStrings(data)

	} else {

		return serializeNumeric(data)

	}
}

func serializeNumeric(data []DataPoint) []byte {

	switch data[0].Value.(type) {
	case float64:
		buffer := make([]byte, len(data)*12)

		for indx, dataPoint := range data {
			binary.LittleEndian.PutUint32(buffer[indx*12:indx*12+4], dataPoint.Timestamp)
			binary.LittleEndian.PutUint64(buffer[indx*12+4:indx*12+12], math.Float64bits(dataPoint.Value.(float64)))
		}

		return buffer
	case float32:
		buffer := make([]byte, len(data)*8)

		for indx, dataPoint := range data {
			binary.LittleEndian.PutUint32(buffer[indx*8:indx*8+4], dataPoint.Timestamp)
			binary.LittleEndian.PutUint32(buffer[indx*8+4:indx*8+8], math.Float32bits(dataPoint.Value.(float32)))
		}

		return buffer

	case int64, int:
		buffer := make([]byte, len(data)*12)

		for indx, dataPoint := range data {
			binary.LittleEndian.PutUint32(buffer[indx*12:indx*12+4], dataPoint.Timestamp)
			binary.LittleEndian.PutUint64(buffer[indx*12+4:indx*12+12], uint64(dataPoint.Value.(int64)))

		}

		return buffer
	case int32:
		buffer := make([]byte, len(data)*8)

		for indx, dataPoint := range data {
			binary.LittleEndian.PutUint32(buffer[indx*8:indx*8+4], dataPoint.Timestamp)
			binary.LittleEndian.PutUint32(buffer[indx*8+4:indx*8+8], uint32(dataPoint.Value.(int32)))
		}

		return buffer
	case uint64, uint:
		buffer := make([]byte, len(data)*12)

		for indx, dataPoint := range data {
			binary.LittleEndian.PutUint32(buffer[indx*12:indx*12+4], dataPoint.Timestamp)
			binary.LittleEndian.PutUint64(buffer[indx*12+4:indx*12+12], uint64(dataPoint.Value.(uint64)))

		}

		return buffer

	case uint32:
		buffer := make([]byte, len(data)*8)

		for indx, dataPoint := range data {
			binary.LittleEndian.PutUint32(buffer[indx*8:indx*8+4], dataPoint.Timestamp)
			binary.LittleEndian.PutUint32(buffer[indx*8+4:indx*8+8], uint32(dataPoint.Value.(uint32)))
		}

		return buffer
	}
	return nil
}

func serializeStrings(data []DataPoint) []byte {
	// Serialize string

	bufferSize := 0

	for _, value := range data {

		bufferSize += len(value.Value.(string)) + 8

	}

	buffer := make([]byte, bufferSize)

	offset := 0
	for _, value := range data {

		binary.LittleEndian.PutUint32(buffer[offset:offset+4], value.Timestamp)

		val := value.Value.(string)

		binary.LittleEndian.PutUint32(buffer[offset+4:offset+8], uint32(len(val)))

		copy(buffer[offset+8:offset+8+len(val)], val)

	}
	return buffer
}

func DeserializeBatch(data []byte, counterId uint16) ([]DataPoint, error) {

	if len(data) == 0 {

		return nil, nil

	}

}
