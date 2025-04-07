package containers

import (
	"fmt"
	"log"
	"os"
	. "reportdb/config"
	"strconv"
	"sync"
	"syscall"
)

type FileMapping struct {
	mapping []byte

	file *os.File

	lock sync.RWMutex

	// Add functionality of access count to periodically unmap less used files.
}

func loadFileMapping(partitionId uint32, storagePath string) (*FileMapping, error) {
	filePath := storagePath + "/data_" + strconv.Itoa(int(partitionId)) + ".bin"

	file, err := os.OpenFile(filePath, os.O_RDWR, 0655)

	if err != nil {

		return nil, err

	}

	fileStats, err := file.Stat()

	if err != nil {

		return nil, err

	}

	fileMapping, err := syscall.Mmap(int(file.Fd()), 0, int(fileStats.Size()), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {

		fmt.Println(err)

		return nil, err

	}

	return &FileMapping{

		mapping: fileMapping,

		file: file,
	}, nil

}

func (fileMapping *FileMapping) UnmapFile() error {
	fileMapping.lock.Lock()

	defer fileMapping.lock.Unlock()

	err := syscall.Munmap(fileMapping.mapping)

	if err != nil {

		log.Println("Error unmapping file", err)

		return err

	}

	err = fileMapping.file.Close()

	if err != nil {

		log.Println("Error closing file", err)

		return err

	}

	return nil

}

func truncateFile(fileMapping *FileMapping) error {

	newSize := int64(len(fileMapping.mapping)) + FileSizeGrowthDelta

	if err := os.Truncate(fileMapping.file.Name(), newSize); err != nil {

		log.Println("Error truncating file", fileMapping.file.Name(), err)

		return err
	}

	// remap the mapping.

	if err := syscall.Munmap(fileMapping.mapping); err != nil {

		log.Println("Error unmapping file", fileMapping.file.Name(), err)

		return err
	}

	newMapping, err := syscall.Mmap(int(fileMapping.file.Fd()), 0, int(newSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {

		log.Println("Error creating mapping file", fileMapping.file.Name(), err)

		return err

	}

	fileMapping.mapping = newMapping

	return nil

}

func (fileMapping *FileMapping) WriteAt(data []byte, offset uint64) error {

	fileMapping.lock.Lock()

	defer fileMapping.lock.Unlock()

	// Check if the file size is sufficient
	if int(offset)+len(data) > int(len(fileMapping.mapping)) {

		err := truncateFile(fileMapping)

		if err != nil {

			return err

		}

	}

	copy(fileMapping.mapping[offset:int(offset)+len(data)], data)

	return nil

}

func (fileMapping *FileMapping) ReadBlocks(objectBlocks []ObjectBlock, blockSize uint32) []byte {

	fileMapping.lock.RLock()

	defer fileMapping.lock.RUnlock()

	data := make([]byte, len(objectBlocks)*int(blockSize)) // make the container for the data.

	var currentIndex = 0
	for _, block := range objectBlocks {

		sizeOfBlockData := blockSize - block.RemainingCapacity
		fmt.Println(block.Offset, block.RemainingCapacity, sizeOfBlockData)

		copy(data[currentIndex:], fileMapping.mapping[int(block.Offset):int(block.Offset)+int(sizeOfBlockData)])

		currentIndex += int(sizeOfBlockData)

	}

	return data[:currentIndex]

}

type OpenFilesPool struct {
	pool map[uint32]*FileMapping

	lock sync.Mutex
}

func NewOpenFilesPool() *OpenFilesPool {

	return &OpenFilesPool{pool: make(map[uint32]*FileMapping)}

}

func (pool *OpenFilesPool) GetFileMapping(partitionId uint32, storagePath string) (*FileMapping, error) {

	pool.lock.Lock()

	defer pool.lock.Unlock()

	if mapping, ok := pool.pool[partitionId]; ok {

		return mapping, nil

	} else {
		// File Not present

		// Create new

		mapping, err := loadFileMapping(partitionId, storagePath)

		if err != nil {

			log.Println("Error opening new File for: ", partitionId, err)

			return nil, err

		}

		pool.pool[partitionId] = mapping

		return mapping, nil
	}
}

func (pool *OpenFilesPool) DeleteFileMapping(partitionId uint32) error {

	pool.lock.Lock()

	defer pool.lock.Unlock()

	mapping := pool.pool[partitionId]

	err := mapping.UnmapFile()

	if err != nil {

		return err

	}

	delete(pool.pool, partitionId)

	return nil
}
