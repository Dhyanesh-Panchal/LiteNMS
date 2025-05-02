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

	queryReceiveChannel := make(chan Query, 10)

	queryResultChannel := make(chan Result, 10)

	storagePool := containers.NewOpenStoragePool()

	var shutdownWaitGroup sync.WaitGroup

	shutdownWaitGroup.Add(1)

	go InitQueryEngine(queryReceiveChannel, queryResultChannel, storagePool, &shutdownWaitGroup)

	query := Query{
		QueryId:               10,
		CounterId:             2,
		From:                  1746037914,
		To:                    1746041514,
		ObjectIds:             []uint32{2130706433},
		VerticalAggregation:   "none",
		HorizontalAggregation: "none",
		Interval:              0,
	}

	queryReceiveChannel <- query

	fmt.Println(<-queryResultChannel)

}
