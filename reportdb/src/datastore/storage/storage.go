package storage

import (
	. "datastore/storage/containers"
	. "datastore/utils"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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

		fmt.Printf("Creating storage directory: %s\n", storagePath)

		err = os.MkdirAll(storagePath, 0755)

		if err != nil {

			log.Println("Failed to create storage directory:", err)

			return err

		}

		// Make partition files and respective index
		for partitionIndex := range partitionCount {

			file, err := os.Create(storagePath + "/data_" + strconv.Itoa(int(partitionIndex)) + ".bin")

			if err != nil {

				log.Println("Error creating new data partition", err)

				return err

			}

			defer func(file *os.File) {

				err := file.Close()

				if err != nil {

					log.Println("Error closing data partition", err)

				}

			}(file)

			err = os.Truncate(file.Name(), InitialFileSize)

			if err != nil {

				log.Println("Error truncating new data partition", err)

				return err

			}

			err = writeNewIndex(storagePath, partitionIndex, blockSize)

		}

	} else if err != nil {

		log.Println("Failed to stat storage directory:", err)

		return err
	}

	return nil
}

func writeNewIndex(storagePath string, partitionIndex uint32, blockSize uint32) error {

	indexFile, err := os.Create(storagePath + "/index_" + strconv.Itoa(int(partitionIndex)) + ".json")

	if err != nil {

		log.Println("Error creating new index partition", err)

		return err

	}

	defer func(indexFile *os.File) {

		err := indexFile.Close()

		if err != nil {

			log.Println("Error closing data partition", err)

		}

	}(indexFile)

	index := NewIndex(blockSize)

	indexBytes, err := json.MarshalIndent(index, "", "  ")

	if err != nil {

		log.Println("Error marshalling index ", err)

		return err

	}

	_, err = indexFile.Write(indexBytes)

	if err != nil {

		log.Println("Error writing index ", err)

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

func (storage *Storage) CloseStorage() {

	storage.openFilesPool.Close()

	storage.indexPool.Close(storage.storagePath)

}
