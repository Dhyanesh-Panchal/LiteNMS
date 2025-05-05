package storage

import (
	. "datastore/storage/containers"
	. "datastore/utils"
	"errors"
	"go.uber.org/zap"
	"os"
	"strconv"
)

var ErrObjectDoesNotExist = errors.New("object does not exist")

var ErrStorageDoesNotExist = errors.New("storage does not exist")

type Storage struct {
	storagePath string

	partitionCount uint32

	blockSize uint32

	openFilesPool *OpenFilesPool

	indexPool *IndexPool
}

func NewStorage(storagePath string, partitionCount uint32, blockSize uint32, createIfNotExist bool) (*Storage, error) {

	// Ensure that storage directory exist, if not create the storage dir and files

	err := ensureStorageDirectory(storagePath, partitionCount, blockSize, createIfNotExist)

	if err != nil {

		return nil, err

	}

	openFilesPool := NewOpenFilesPool()

	indexPool := NewIndexPool()

	return &Storage{
		storagePath,
		partitionCount,
		blockSize,
		openFilesPool,
		indexPool,
	}, nil
}

func ensureStorageDirectory(storagePath string, partitionCount uint32, blockSize uint32, createIfNotExist bool) error {

	_, err := os.Stat(storagePath)

	if os.IsNotExist(err) {

		if !createIfNotExist {

			return ErrStorageDoesNotExist

		}

		Logger.Info("Creating storage", zap.String("storagePath", storagePath))

		err = os.MkdirAll(storagePath, 0755)

		if err != nil {

			Logger.Info("Failed to create storage directory:", zap.Error(err))

			return err

		}

		// Make partition files and respective index
		for partitionIndex := range partitionCount {

			file, err := os.Create(storagePath + "/data_" + strconv.Itoa(int(partitionIndex)) + ".bin")

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

			err = os.Truncate(file.Name(), InitialFileSize)

			if err != nil {

				Logger.Error("error truncating new data partition", zap.Error(err))

				return err

			}

			index := NewIndex(blockSize)

			err = index.WriteIndexToFile(storagePath, partitionIndex)

			if err != nil {

				Logger.Error("error marshalling index ", zap.Error(err))

				return err

			}
		}

	} else if err != nil {

		Logger.Info("Failed to stat storage directory:", zap.Error(err))

		return err
	}

	return nil
}

//func writeNewIndex(storagePath string, partitionIndex uint32, blockSize uint32) error {
//
//	indexFile, err := os.Create(storagePath + "/index_" + strconv.Itoa(int(partitionIndex)) + ".bin")
//
//	if err != nil {
//
//		Logger.Error("error creating new index partition", zap.Error(err))
//
//		return err
//
//	}
//
//	defer func(indexFile *os.File) {
//
//		err := indexFile.Close()
//
//		if err != nil {
//
//			Logger.Error("error closing data partition", zap.Error(err))
//
//		}
//
//	}(indexFile)
//
//
//
//
//	return nil
//}

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

	err = DiskWrite(key, value, file, index)

	if err != nil {

		return err

	}

	err = index.WriteIndexToFile(storage.storagePath, key%storage.partitionCount)

	if err != nil {

		return err

	}

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

func (storage *Storage) CloseStorage() {

	storage.openFilesPool.Close()

	storage.indexPool.Close(storage.storagePath)

}
