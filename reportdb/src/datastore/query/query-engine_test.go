package query

import (
	"datastore/containers"
	"datastore/utils"
	"fmt"
	"sync"
	"testing"
)

func TestQueryEngine(t *testing.T) {

	utils.LoadConfig()

	queryReceiveChannel := make(chan map[string]interface{}, 10)

	queryResultChannel := make(chan Result, 10)

	storagePool := containers.NewOpenStoragePool()

	var shutdownWaitGroup sync.WaitGroup

	shutdownWaitGroup.Add(1)

	go InitQueryEngine(queryReceiveChannel, queryResultChannel, storagePool, &shutdownWaitGroup)

	query := map[string]interface{}{
		"queryId":               float64(10),
		"counterId":             float64(2),
		"from":                  float64(1744716600),
		"to":                    float64(1744867800),
		"objectIds":             []uint32{169093219, 2130706433},
		"verticalAggregation":   "sum",
		"horizontalAggregation": "avg",
		"interval":              float64(0),
	}

	queryReceiveChannel <- query

	fmt.Println(<-queryResultChannel)

}
