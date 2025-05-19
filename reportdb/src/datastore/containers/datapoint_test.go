package containers

import (
	"fmt"
	"testing"
)

func TestDeserializeBatch(t *testing.T) {

	data := []DataPoint{
		{
			1744089990, 12,
		},
		{
			1744089990, 5,
		},
	}

	serialData := make([]byte, 0)

	_ = SerializeBatch(data, &serialData, "int64")

	fmt.Println(serialData)

	newData, err := DeserializeBatch(serialData, "int64")

	if err != nil {

		t.Error(err)

	}

	fmt.Println(newData)

}
