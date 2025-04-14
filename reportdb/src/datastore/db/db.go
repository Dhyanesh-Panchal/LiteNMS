package db

import (
	. "datastore/containers"
	. "datastore/reader"
	. "datastore/utils"
	. "datastore/writer"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type ReportDB struct {
	storagePool *StoragePool

	dataWriteChannel chan []PolledDataPoint
}

func InitDB(dataWriteChannel <-chan []PolledDataPoint, queryReceiveChannel <-chan Query, queryResultChannel chan<- Result, globalShutdown <-chan bool, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	defer log.Println("Database closed")

	// Ensure storage directory is created.
	err := os.MkdirAll(filepath.Dir(filepath.Dir(CurrentWorkingDirectory))+"/data", 0777)

	if err != nil {

		log.Println("Error creating data directory:", err)

		return

	}

	storagePool := NewOpenStoragePool()

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

//func (db ReportDB) QueryHistogram(from uint32, to uint32, counterId uint16, objects []uint32) (map[uint32][]DataPoint, error) {
//
//	finalData := map[uint32][]DataPoint{}
//
//	for date := from - (from % 86400); date <= to; date += 86400 {
//
//		dateObject := UnixToDate(date)
//
//		storageKey := StoragePoolKey{
//
//			Date: dateObject,
//
//			CounterId: counterId,
//		}
//
//		storageEngine, err := db.storagePool.GetStorage(storageKey, false)
//
//		if err != nil {
//
//			if errors.Is(err, ErrStorageDoesNotExist) {
//
//				continue
//
//			}
//
//			return nil, err
//
//		}
//
//		readSingleDay(dateObject, storageEngine, counterId, objects, finalData, from, to)
//
//		db.storagePool.CloseStorage(storageKey)
//
//	}
//
//	return finalData, nil
//
//}
