package containers

import (
	"encoding/binary"
	"math"
	"reflect"
)

type DataPoint struct {
	Timestamp uint32
	Value     interface{}
}

func (dataPoint *DataPoint) Serialize(totalDataPointSize uint32) []byte {

	buffer := make([]byte, totalDataPointSize)

	binary.LittleEndian.PutUint32(buffer[:4], dataPoint.Timestamp)

	dataType := reflect.TypeOf(dataPoint.Value).Kind()

	switch dataType {

	case reflect.Float64:

		binary.LittleEndian.PutUint64(buffer[4:], math.Float64bits(dataPoint.Value.(float64)))

	case reflect.Float32:

		binary.LittleEndian.PutUint32(buffer[4:], math.Float32bits(dataPoint.Value.(float32)))

	case reflect.Int64:

		binary.LittleEndian.PutUint64(buffer[4:], uint64(dataPoint.Value.(int64)))

	case reflect.Int32:

		binary.LittleEndian.PutUint32(buffer[4:], uint32(dataPoint.Value.(int32)))

	case reflect.Uint64:

		binary.LittleEndian.PutUint64(buffer[4:], dataPoint.Value.(uint64))

	case reflect.Uint32:

		binary.LittleEndian.PutUint32(buffer[4:], dataPoint.Value.(uint32))

	}

	return buffer

}
