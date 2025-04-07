package reportdb

import (
	. "reportdb/containers"
	"reportdb/writer"
)

type ReportDB struct {
	storagePool *StoragePool

	dataWriteChannel chan []PolledDataPoint
}

func InitDB() ReportDB {

	storagePool := NewOpenStoragePool()

	dataWriteChannel := make(chan []PolledDataPoint)

	go writer.InitWriter(dataWriteChannel, storagePool)

	return ReportDB{
		storagePool:      storagePool,
		dataWriteChannel: dataWriteChannel,
	}

}

func (db ReportDB) Write(records []PolledDataPoint) {

	db.dataWriteChannel <- records

}
