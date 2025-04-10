package containers

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

type DataPoint struct {
	Timestamp uint32
	Value     interface{}
}

func SerializeBatch(data []DataPoint, dataType string) ([]byte, error) {

	if len(data) == 0 {

		return nil, nil

	}

	switch dataType {
	case "float64":
		return serializeFloat64(data), nil
	case "float32":
		return serializeFloat32(data), nil
	case "uint64", "uint", "int64", "int":
		return serializeUint64(data), nil
	case "uint32", "int32":
		return serializeUint32(data), nil
	case "string":
		return serializeStrings(data), nil
	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}
}

func serializeFloat64(data []DataPoint) []byte {

	buffer := make([]byte, len(data)*12)

	for index, dataPoint := range data {

		binary.LittleEndian.PutUint32(buffer[index*12:index*12+4], dataPoint.Timestamp)

		binary.LittleEndian.PutUint64(buffer[index*12+4:index*12+12], math.Float64bits(dataPoint.Value.(float64)))

	}

	return buffer
}

func serializeFloat32(data []DataPoint) []byte {

	buffer := make([]byte, len(data)*8)

	for index, dataPoint := range data {

		binary.LittleEndian.PutUint32(buffer[index*8:index*8+4], dataPoint.Timestamp)

		binary.LittleEndian.PutUint32(buffer[index*8+4:index*8+8], math.Float32bits(dataPoint.Value.(float32)))

	}

	return buffer
}

//func serializeInt64(data []DataPoint) []byte {
//	buffer := make([]byte, len(data)*12)
//
//	for index, dataPoint := range data {
//		binary.LittleEndian.PutUint32(buffer[index*12:index*12+4], dataPoint.Timestamp)
//		binary.LittleEndian.PutUint64(buffer[index*12+4:index*12+12], uint64(dataPoint.Value.(float64)))
//
//	}
//
//	return buffer
//}
//func serializeInt32(data []DataPoint) []byte {
//	buffer := make([]byte, len(data)*8)
//
//	for index, dataPoint := range data {
//		binary.LittleEndian.PutUint32(buffer[index*8:index*8+4], dataPoint.Timestamp)
//		binary.LittleEndian.PutUint32(buffer[index*8+4:index*8+8], uint32(dataPoint.Value.(float64)))
//	}
//
//	return buffer
//}

func serializeUint64(data []DataPoint) []byte {

	buffer := make([]byte, len(data)*12)

	for index, dataPoint := range data {

		binary.LittleEndian.PutUint32(buffer[index*12:index*12+4], dataPoint.Timestamp)

		binary.LittleEndian.PutUint64(buffer[index*12+4:index*12+12], uint64(dataPoint.Value.(float64)))

	}

	return buffer
}

func serializeUint32(data []DataPoint) []byte {

	buffer := make([]byte, len(data)*8)

	for index, dataPoint := range data {

		binary.LittleEndian.PutUint32(buffer[index*8:index*8+4], dataPoint.Timestamp)

		binary.LittleEndian.PutUint32(buffer[index*8+4:index*8+8], uint32(dataPoint.Value.(float64)))

	}

	return buffer
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

		offset += 8 + len(val)

	}
	return buffer
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
