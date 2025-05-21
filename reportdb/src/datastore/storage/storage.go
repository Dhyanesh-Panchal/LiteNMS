package storage

import (
	. "datastore/storage/containers"
	. "datastore/utils"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"os"
	"strconv"
	"time"
)

var ErrObjectDoesNotExist = errors.New("object does not exist")

var ErrStorageDoesNotExist = errors.New("storage does not exist")

type Storage struct {
	storagePath string

	partitionCount uint32

	blockSize uint32

	openFilesPool *OpenFilesPool

	indexPool *IndexPool

	indexSyncTicker *time.Ticker

	syncRoutineShutdown chan struct{}
}

func NewStorage(storagePath string, partitionCount uint32, blockSize uint32, createIfNotExist bool) (*Storage, error) {

	// Ensure that storage directory exist, if not create the storage dir and files

	if err := ensureStorageDirectory(storagePath, partitionCount, blockSize, createIfNotExist); err != nil {

		return nil, err

	}

	storage := &Storage{

		storagePath: storagePath,

		partitionCount: partitionCount,

		blockSize: blockSize,

		openFilesPool: NewOpenFilesPool(),

		indexPool: NewIndexPool(),

		indexSyncTicker: time.NewTicker(time.Second * 7),

		syncRoutineShutdown: make(chan struct{}, 1),
	}

	go indexSyncRoutine(storage, storage.syncRoutineShutdown)

	return storage, nil
}

func indexSyncRoutine(storage *Storage, syncRoutineShutdown chan struct{}) {

	for {

		select {

		case <-syncRoutineShutdown:

			storage.indexSyncTicker.Stop()

			return

		case <-storage.indexSyncTicker.C:

			storage.indexPool.Sync(storage.storagePath)

		}
	}

}

func ensureStorageDirectory(storagePath string, partitionCount uint32, blockSize uint32, createIfNotExist bool) error {

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {

		if !createIfNotExist {

			return ErrStorageDoesNotExist

		}

		Logger.Info("Creating storage", zap.String("storagePath", storagePath))

		if err = os.MkdirAll(storagePath, 0755); err != nil {

			Logger.Error("Failed to create storage directory:", zap.Error(err))

			return err

		}

		// Make partition files and respective index
		for partitionId := range partitionCount {

			file, err := os.Create(storagePath + "/data_" + strconv.Itoa(int(partitionId)) + ".bin")

			if err != nil {

				Logger.Error("error creating new data partition", zap.Error(err))

				return err

			}

			defer func(file *os.File) {

				err := file.Close()

				if err != nil {

					Logger.Error("error closing data partition", zap.Error(err))

				}

			}(file)

			if err = os.Truncate(file.Name(), InitialFileSize); err != nil {

				Logger.Error("error truncating new data partition", zap.Error(err))

				return err

			}

			index := NewIndex(blockSize)

			indexBytes, err := msgpack.Marshal(index)

			indexFilePath := storagePath + "/index_" + strconv.Itoa(int(partitionId)) + ".bin"

			err = os.WriteFile(indexFilePath, indexBytes, 0644)

			if err != nil {

				Logger.Error("error creating index file", zap.Error(err))

				return err

			}

		}

	} else if err != nil {

		Logger.Error("Failed to stat storage directory:", zap.Error(err))

		return err
	}

	return nil
}

// -------------- Storage Engine Interface functions -----------------

func (storage *Storage) Put(key uint32, value []byte) error {

	file, err := storage.openFilesPool.GetFileMapping(key%storage.partitionCount, storage.storagePath)

	if err != nil {

		return err

	}

	index, err := storage.indexPool.Get(key%storage.partitionCount, storage.storagePath)

	if err != nil {

		return err

	}

	if err = DiskWrite(key, value, file, index); err != nil {

		return err

	}

	//if err = index.syncFile(storage.storagePath, key%storage.partitionCount); err != nil {
	//
	//	return err
	//
	//}

	return nil

}

func (storage *Storage) Get(key uint32) ([]byte, error) {

	file, err := storage.openFilesPool.GetFileMapping(key%storage.partitionCount, storage.storagePath)

	if err != nil {

		return nil, err

	}

	index, err := storage.indexPool.Get(key%storage.partitionCount, storage.storagePath)

	if err != nil {

		return nil, err

	}

	blocks := index.GetIndexObjectBlocks(key)

	if blocks == nil {

		return nil, ErrObjectDoesNotExist

	}

	data := file.ReadBlocks(blocks, storage.blockSize)

	return data, nil

}

func (storage *Storage) GetAllKeys() ([]uint32, error) {

	keys := make([]uint32, 0)

	// Get all keys from all partitions
	for partitionIndex := range storage.partitionCount {

		index, err := storage.indexPool.Get(partitionIndex, storage.storagePath)

		if err != nil {

			return nil, err

		}

		for key := range index.ObjectIndex {

			keys = append(keys, key)

		}

	}

	return keys, nil

}

func (storage *Storage) Close() {

	storage.syncRoutineShutdown <- struct{}{}

	storage.openFilesPool.Close()

	storage.indexPool.Close(storage.storagePath)

}
