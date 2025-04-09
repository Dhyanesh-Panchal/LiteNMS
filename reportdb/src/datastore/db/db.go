package db

import (
	. "datastore/containers"
	. "datastore/reader"
	. "datastore/utils"
	. "datastore/writer"
	"log"
	"os"
	"path/filepath"
	"time"
)

type ReportDB struct {
	storagePool *StoragePool

	dataWriteChannel chan []PolledDataPoint
}

func InitDB() (*ReportDB, error) {

	// Ensure storage directory is created.
	err := os.MkdirAll(filepath.Dir(filepath.Dir(CurrentWorkingDirectory))+"/data", 0777)

	if err != nil {

		log.Println(err)

		return nil, err

	}

	storagePool := NewOpenStoragePool()

	dataWriteChannel := make(chan []PolledDataPoint)

	go InitWriter(dataWriteChannel, storagePool)

	return &ReportDB{
		storagePool:      storagePool,
		dataWriteChannel: dataWriteChannel,
	}, nil

}

func (db ReportDB) Write(records []PolledDataPoint) {

	db.dataWriteChannel <- records

}

func (db ReportDB) QueryHistogram(from uint32, to uint32, counterId uint16, objects []uint32) (map[uint32][]DataPoint, error) {

	finalData := map[uint32][]DataPoint{}

	for date := time.Unix(int64(from), 0).UTC(); date.Before(time.Unix(int64(to), 0)) || date.Equal(time.Unix(int64(to), 0)); date = date.Add(time.Hour * 24) {

		dateObject := TimeToDate(date)

		storageEngine, err := db.storagePool.GetStorage(StoragePoolKey{
			Date:      dateObject,
			CounterId: counterId,
		}, false)

		if err != nil {

			return nil, err

		}

		ReadFullDate(dateObject, storageEngine, counterId, objects, finalData, from, to)

	}

	return finalData, nil

}
