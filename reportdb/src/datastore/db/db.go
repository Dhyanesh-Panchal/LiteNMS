package db

import (
	. "datastore/containers"
	. "datastore/reader"
	. "datastore/storage"
	. "datastore/utils"
	. "datastore/writer"
	"errors"
	"log"
	"os"
	"path/filepath"
)

type ReportDB struct {
	storagePool *StoragePool

	dataWriteChannel chan []PolledDataPoint
}

func InitDB(dataWriteChannel chan []PolledDataPoint) (*ReportDB, error) {

	// Ensure storage directory is created.
	err := os.MkdirAll(filepath.Dir(filepath.Dir(CurrentWorkingDirectory))+"/data", 0777)

	if err != nil {

		log.Println(err)

		return nil, err

	}

	storagePool := NewOpenStoragePool()

	go InitWriteHandler(dataWriteChannel, storagePool)

	return &ReportDB{

		storagePool: storagePool,

		dataWriteChannel: dataWriteChannel,
	}, nil

}

func (db ReportDB) Write(records []PolledDataPoint) {

	db.dataWriteChannel <- records

}

func (db ReportDB) QueryHistogram(from uint32, to uint32, counterId uint16, objects []uint32) (map[uint32][]DataPoint, error) {

	finalData := map[uint32][]DataPoint{}

	for date := from; date <= to; date += 86400 {

		dateObject := UnixToDate(date)

		storageKey := StoragePoolKey{

			Date: dateObject,

			CounterId: counterId,
		}

		storageEngine, err := db.storagePool.GetStorage(storageKey, false)

		if err != nil {

			if errors.Is(err, ErrStorageDoesNotExist) {

				continue

			}

			return nil, err

		}

		ReadFullDate(dateObject, storageEngine, counterId, objects, finalData, from, to)

		db.storagePool.CloseStorage(storageKey)

	}

	return finalData, nil

}
