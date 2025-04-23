package query

import (
	. "datastore/containers"
	. "datastore/utils"
	"go.uber.org/zap"
	"reflect"
	"sync"
)

func VerticalAggregator(daysData []map[uint32][]DataPoint, aggregation string, dataType string, numberOfObjects int) {

	var completionWg sync.WaitGroup

	for index := range len(daysData) {

		completionWg.Add(1)

		go verticalAggregateSingleDay(daysData, index, aggregation, dataType, numberOfObjects, &completionWg)

	}

	completionWg.Wait()

}

func verticalAggregateSingleDay(daysData []map[uint32][]DataPoint, dayIndex int, aggregation string, dataType string, numberOfObjects int, completionWg *sync.WaitGroup) {

	defer completionWg.Done()

	day := daysData[dayIndex]

	timeIndexedAggregatedData := make(map[uint32]interface{})

	for objectId, points := range day {

		for _, point := range points {

			if val, exist := timeIndexedAggregatedData[point.Timestamp]; exist {

				timeIndexedAggregatedData[point.Timestamp] = aggregatePoint(val, point.Value, aggregation, dataType)

			}

		}

		delete(day, objectId)

	}

	if aggregation == "avg" {

		for _, val := range timeIndexedAggregatedData {

			val = Divide(val, numberOfObjects)

		}

	}

}

func aggregatePoint(value1 interface{}, value2 interface{}, aggregation string, dataType string) interface{} {

	switch aggregation {

	case "max":

		return Max(value1, value2)

	case "min":

		return Min(value1, value2)

	case "sum", "avg":

		return Sum(value1, value2)

	}

	Logger.Info("aggregation not supported", zap.String("aggregation", aggregation))

	return nil
}

func Max(value1 interface{}, value2 interface{}, dataType string) interface{} {

	switch dataType {

	case "float64":
		if value1.(float64) > value2.(float64) {

			return value1

		} else {

			return value2

		}

	case "float32":
		if value1.(float32) > value2.(float32) {

			return value1

		} else {

			return value2

		}

	}

	Logger.Info("aggregation not supported", zap.String("aggregation", dataType.String()))
	return nil

}
