package db

import (
	. "datastore/containers"
	. "datastore/query"
	. "datastore/utils"
	. "datastore/writer"
	"go.uber.org/zap"
	"os"
	"sync"
)

type ReportDB struct {
	storagePool *StoragePool

	dataWriteChannel chan []PolledDataPoint
}

func InitDB(dataWriteChannel <-chan []PolledDataPoint, queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, globalShutdown <-chan bool, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	defer Logger.Info("database closed")

	// Ensure storage directory is created.

	if err := os.MkdirAll(StorageDirectory, 0777); err != nil {

		Logger.Fatal("error creating data directory:" + err.Error())

		return

	}

	// Initialize Containers

	storagePool := InitStoragePool()

	if err := InitDataPointsCache(); err != nil {

		Logger.Error("Error initializing data points cache", zap.Error(err))

	}

	var dbShutdownWaitGroup sync.WaitGroup

	dbShutdownWaitGroup.Add(2)

	go InitWriteHandler(dataWriteChannel, storagePool, &dbShutdownWaitGroup)

	go InitQueryEngine(queryReceiveChannel, queryResultChannel, storagePool, &dbShutdownWaitGroup)

	<-globalShutdown

	// Wait for writer Reader to shut down
	dbShutdownWaitGroup.Wait()

	// Close the storagePool
	storagePool.ClosePool()

}
