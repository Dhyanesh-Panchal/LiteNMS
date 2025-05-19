package containers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

type DataPoint struct {
	Timestamp uint32 `json:"timestamp" msgpack:"timestamp"`

	Value interface{} `json:"value" msgpack:"value"`
}

func SerializeBatch(data []DataPoint, dataContainer *[]byte, dataType string) error {

	if len(data) == 0 {

		return nil

	}

	switch dataType {

	case "float64":

		serializeFloat64(data, dataContainer)

		return nil

	case "float32":

		serializeFloat32(data, dataContainer)

		return nil

	case "uint64", "uint", "int64", "int":

		serializeUint64(data, dataContainer)

		return nil

	case "uint32", "int32":

		serializeUint32(data, dataContainer)

		return nil

	case "string":

		serializeStrings(data, dataContainer)

		return nil
	default:
		return fmt.Errorf("unsupported data type: %s", dataType)
	}
}

func serializeFloat64(data []DataPoint, dataContainer *[]byte) {

	if cap(*dataContainer) < len(data)*12 {

		*dataContainer = make([]byte, len(data)*12)

	} else {

		*dataContainer = (*dataContainer)[:len(data)*12]

	}

	for index, dataPoint := range data {

		binary.LittleEndian.PutUint32((*dataContainer)[index*12:index*12+4], dataPoint.Timestamp)

		binary.LittleEndian.PutUint64((*dataContainer)[index*12+4:index*12+12], math.Float64bits(dataPoint.Value.(float64)))

	}
}

func serializeFloat32(data []DataPoint, dataContainer *[]byte) {

	if cap(*dataContainer) < len(data)*8 {

		*dataContainer = make([]byte, len(data)*8)

	} else {

		*dataContainer = (*dataContainer)[:len(data)*8]

	}

	for index, dataPoint := range data {

		binary.LittleEndian.PutUint32((*dataContainer)[index*8:index*8+4], dataPoint.Timestamp)

		binary.LittleEndian.PutUint32((*dataContainer)[index*8+4:index*8+8], math.Float32bits(dataPoint.Value.(float32)))

	}
}

func serializeUint64(data []DataPoint, dataContainer *[]byte) {

	if cap(*dataContainer) < len(data)*12 {

		*dataContainer = make([]byte, len(data)*12)

	} else {

		*dataContainer = (*dataContainer)[:len(data)*12]

	}

	for index, dataPoint := range data {

		binary.LittleEndian.PutUint32((*dataContainer)[index*12:index*12+4], dataPoint.Timestamp)

		binary.LittleEndian.PutUint64((*dataContainer)[index*12+4:index*12+12], uint64(dataPoint.Value.(float64)))

	}
}

func serializeUint32(data []DataPoint, dataContainer *[]byte) {

	if cap(*dataContainer) < len(data)*8 {

		*dataContainer = make([]byte, len(data)*8)

	} else {

		*dataContainer = (*dataContainer)[:len(data)*8]

	}

	for index, dataPoint := range data {

		binary.LittleEndian.PutUint32((*dataContainer)[index*8:index*8+4], dataPoint.Timestamp)

		binary.LittleEndian.PutUint32((*dataContainer)[index*8+4:index*8+8], uint32(dataPoint.Value.(float64)))

	}

}

func serializeStrings(data []DataPoint, dataContainer *[]byte) {
	// Serialize string

	bufferSize := 0

	for _, value := range data {

		bufferSize += len(value.Value.(string)) + 8

	}

	if cap(*dataContainer) < bufferSize {

		*dataContainer = make([]byte, bufferSize)

	} else {

		*dataContainer = (*dataContainer)[:bufferSize]

	}

	offset := 0
	for _, value := range data {

		binary.LittleEndian.PutUint32((*dataContainer)[offset:offset+4], value.Timestamp)

		val := value.Value.(string)

		binary.LittleEndian.PutUint32((*dataContainer)[offset+4:offset+8], uint32(len(val)))

		copy((*dataContainer)[offset+8:offset+8+len(val)], val)

		offset += 8 + len(val)

	}
}

// --------------- Deserialize-------------

func DeserializeBatch(data []byte, dataType string) ([]DataPoint, error) {

	if len(data) == 0 {
		return nil, nil
	}

	switch dataType {

	case "float64":
		return deserializeFloat64(data)

	case "float32":
		return deserializeFloat32(data)

	case "int64", "int":
		return deserializeInt64(data)

	case "int32":
		return deserializeInt32(data)

	case "uint64", "uint":
		return deserializeUint64(data)

	case "uint32":
		return deserializeUint32(data)

	case "string":
		return deserializeStrings(data)

	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)

	}

}

func deserializeFloat64(data []byte) ([]DataPoint, error) {

	if len(data)%12 != 0 {
		return nil, errors.New("invalid data length for float64")
	}

	count := len(data) / 12

	points := make([]DataPoint, count)

	for i := 0; i < count; i++ {

		timestamp := binary.LittleEndian.Uint32(data[i*12 : i*12+4])

		valueBits := binary.LittleEndian.Uint64(data[i*12+4 : i*12+12])

		points[i] = DataPoint{

			Timestamp: timestamp,

			Value: math.Float64frombits(valueBits),
		}

	}

	return points, nil

}

func deserializeFloat32(data []byte) ([]DataPoint, error) {

	if len(data)%8 != 0 {
		return nil, errors.New("invalid data length for float32")
	}

	count := len(data) / 8

	points := make([]DataPoint, count)

	for i := 0; i < count; i++ {

		timestamp := binary.LittleEndian.Uint32(data[i*8 : i*8+4])

		valueBits := binary.LittleEndian.Uint32(data[i*8+4 : i*8+8])

		points[i] = DataPoint{

			Timestamp: timestamp,

			Value: math.Float32frombits(valueBits),
		}

	}

	return points, nil

}

func deserializeInt64(data []byte) ([]DataPoint, error) {

	if len(data)%12 != 0 {
		return nil, errors.New("invalid data length for int64")
	}

	count := len(data) / 12

	points := make([]DataPoint, count)

	for i := 0; i < count; i++ {

		timestamp := binary.LittleEndian.Uint32(data[i*12 : i*12+4])

		value := int64(binary.LittleEndian.Uint64(data[i*12+4 : i*12+12]))

		points[i] = DataPoint{Timestamp: timestamp, Value: value}

	}

	return points, nil

}

func deserializeInt32(data []byte) ([]DataPoint, error) {

	if len(data)%8 != 0 {
		return nil, errors.New("invalid data length for int32")
	}

	count := len(data) / 8

	points := make([]DataPoint, count)

	for i := 0; i < count; i++ {

		timestamp := binary.LittleEndian.Uint32(data[i*8 : i*8+4])

		value := int32(binary.LittleEndian.Uint32(data[i*8+4 : i*8+8]))

		points[i] = DataPoint{Timestamp: timestamp, Value: value}

	}

	return points, nil

}

func deserializeUint64(data []byte) ([]DataPoint, error) {

	if len(data)%12 != 0 {
		return nil, errors.New("invalid data length for uint64")
	}

	count := len(data) / 12

	points := make([]DataPoint, count)

	for i := 0; i < count; i++ {

		timestamp := binary.LittleEndian.Uint32(data[i*12 : i*12+4])

		value := binary.LittleEndian.Uint64(data[i*12+4 : i*12+12])

		points[i] = DataPoint{Timestamp: timestamp, Value: value}

	}

	return points, nil

}

func deserializeUint32(data []byte) ([]DataPoint, error) {

	if len(data)%8 != 0 {
		return nil, errors.New("invalid data length for uint32")
	}

	count := len(data) / 8

	points := make([]DataPoint, count)

	for i := 0; i < count; i++ {

		timestamp := binary.LittleEndian.Uint32(data[i*8 : i*8+4])

		value := binary.LittleEndian.Uint32(data[i*8+4 : i*8+8])

		points[i] = DataPoint{Timestamp: timestamp, Value: value}

	}

	return points, nil

}

func deserializeStrings(data []byte) ([]DataPoint, error) {

	var points []DataPoint

	offset := 0

	for offset < len(data) {

		if offset+8 > len(data) {
			return nil, errors.New("unexpected end of data")
		}

		timestamp := binary.LittleEndian.Uint32(data[offset : offset+4])

		length := binary.LittleEndian.Uint32(data[offset+4 : offset+8])

		if offset+8+int(length) > len(data) {
			return nil, errors.New("string length goes out of bounds")
		}

		value := string(data[offset+8 : offset+8+int(length)])

		points = append(points, DataPoint{

			Timestamp: timestamp,

			Value: value,
		})

		offset += 8 + int(length)

	}

	return points, nil
}
