package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	. "reportdb/config"
	. "reportdb/storage/containers"
	"strconv"
)

type Storage struct {
	storagePath string

	partitionCount uint32

	blockSize uint32

	openFilesPool *OpenFilesPool

	indexPool *IndexPool
}

var ErrObjectDoesNotExist = errors.New("object does not exist")

func NewStorage(storagePath string, partitionCount uint32, blockSize uint32) (*Storage, error) {

	// Ensure that storage directory exist, if not create the storage dir and files

	err := ensureStorageDirectory(storagePath, partitionCount, blockSize)

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

func ensureStorageDirectory(storagePath string, partitionCount uint32, blockSize uint32) error {

	_, err := os.Stat(storagePath)

	if os.IsNotExist(err) {

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

			defer file.Close()

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
	defer indexFile.Close()

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

func (e *Storage) Put(objectId uint32, value []byte) error {

	file, err := e.openFilesPool.GetFileMapping(objectId%e.partitionCount, e.storagePath)

	if err != nil {

		return err

	}

	index, err := e.indexPool.Get(objectId%e.partitionCount, e.storagePath)

	if err != nil {

		return err

	}

	err = DiskWrite(objectId, value, file, index)

	if err != nil {

		return err

	}

	err = index.WriteIndexToFile(e.storagePath, objectId%e.partitionCount)

	if err != nil {

		return err

	}

	return nil

}

func (e *Storage) Get(objectId uint32) ([]byte, error) {

	file, err := e.openFilesPool.GetFileMapping(objectId%e.partitionCount, e.storagePath)

	if err != nil {

		return nil, err

	}

	index, err := e.indexPool.Get(objectId%e.partitionCount, e.storagePath)

	if err != nil {

		return nil, err

	}

	blocks := index.GetIndexObjectBlocks(objectId)

	if blocks == nil {

		return nil, ErrObjectDoesNotExist

	}

	data := file.ReadBlocks(blocks, e.blockSize)

	return data, nil

}
