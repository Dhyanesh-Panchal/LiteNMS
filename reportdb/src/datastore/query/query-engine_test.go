package query

import (
	"datastore/containers"
	"datastore/utils"
	"fmt"
	"sync"
	"testing"
)

func TestQueryEngine(t *testing.T) {

	_ = utils.LoadConfig()

	_ = utils.InitLogger()

	queryReceiveChannel := make(chan Query, 10)

	queryResultChannel := make(chan Result, 10)

	storagePool := containers.NewOpenStoragePool()

	var shutdownWaitGroup sync.WaitGroup

	shutdownWaitGroup.Add(1)

	go InitQueryEngine(queryReceiveChannel, queryResultChannel, storagePool, &shutdownWaitGroup)

	query := Query{
		QueryId:               10,
		CounterId:             2,
		From:                  1746505800,
		To:                    1746541800,
		ObjectIds:             []uint32{2886731847},
		ObjectWiseAggregation: "none",
		TimestampAggregation:  "none",
		Interval:              0,
	}

	queryReceiveChannel <- query

	result := <-queryResultChannel

	fmt.Println(result.Data[2886731847])

}

func TestQueryEngine2(t *testing.T) {

	_ = utils.LoadConfig()

	_ = utils.InitLogger()

	queryReceiveChannel := make(chan Query, 10)

	queryResultChannel := make(chan Result, 10)

	storagePool := containers.NewOpenStoragePool()

	var shutdownWaitGroup sync.WaitGroup

	shutdownWaitGroup.Add(1)

	go InitQueryEngine(queryReceiveChannel, queryResultChannel, storagePool, &shutdownWaitGroup)

	query := Query{
		QueryId:               1,
		CounterId:             2,
		From:                  1746515997,
		To:                    1746541800,
		ObjectIds:             []uint32{2886731920},
		ObjectWiseAggregation: "none",
		TimestampAggregation:  "none",
		Interval:              0,
	}

	queryReceiveChannel <- query

	result := <-queryResultChannel

	fmt.Println(len(result.Data[2886731920]))

}
