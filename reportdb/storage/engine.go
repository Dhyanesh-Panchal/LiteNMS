package storage

import (
	. "reportdb/config"
	. "reportdb/storage/containers"
	"reportdb/storage/dataput"
)

type Engine struct {
	openFilesPool *OpenFilesPool

	indexPool *IndexPool
}

func NewEngine() *Engine {

	openFilesPool := NewOpenFilesPool()

	indexPool := NewIndexPool()

	return &Engine{
		openFilesPool,
		indexPool,
	}
}

func (e *Engine) Put(key Key, values []DataPoint) error {

	file, err := e.openFilesPool.Get(FilesPoolKey{

		CounterId: key.CounterId,

		PartitionIndex: key.ObjectId % PartitionCount,

		Date: key.Date,
	})

	if err != nil {

		return err

	}

	index, err := e.indexPool.Get(IndexPoolKey{Date: key.Date, CounterId: key.CounterId})

	if err != nil {

		return err

	}

	err = dataput.DiskWrite(values, key, file, index)

	return err

}

//func (e *Engine) Get(key Key) ([]DataPoint, error) {
//
//
//
//}
