package containers

import (
	. "datastore/utils"
	"errors"
	"go.uber.org/zap"
	"os"
	"strconv"
	"sync"
	"syscall"
)

var ErrUnmappingFile = errors.New("error unmapping file")

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

		if os.IsNotExist(err) {

			file, err = os.Create(filePath)

			if err != nil {

				return nil, err

			}

		}

		return nil, err

	}

	fileStats, err := file.Stat()

	if err != nil {

		return nil, err

	}

	fileMapping, err := syscall.Mmap(int(file.Fd()), 0, int(fileStats.Size()), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {

		Logger.Error("error mapping the file for", zap.String("storagePath", storagePath), zap.Uint32("partitionId:", partitionId))

		return nil, err

	}

	return &FileMapping{

		mapping: fileMapping,

		file: file,
	}, nil

}

func truncateFile(fileMapping *FileMapping) error {

	newSize := int64(len(fileMapping.mapping)) + FileSizeGrowthDelta

	if err := os.Truncate(fileMapping.file.Name(), newSize); err != nil {

		Logger.Error("error truncating file", zap.String("FileName", fileMapping.file.Name()), zap.Error(err))

		return err
	}

	// remap the mapping.

	if err := syscall.Munmap(fileMapping.mapping); err != nil {

		Logger.Error(ErrUnmappingFile.Error())

		return ErrUnmappingFile
	}

	newMapping, err := syscall.Mmap(int(fileMapping.file.Fd()), 0, int(newSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {

		Logger.Error("Error creating mapping", zap.String("fileName", fileMapping.file.Name()), zap.Error(err))

		return err

	}

	fileMapping.mapping = newMapping

	return nil

}

func (fileMapping *FileMapping) UnmapFile() error {
	fileMapping.lock.Lock()

	defer fileMapping.lock.Unlock()

	if err := syscall.Munmap(fileMapping.mapping); err != nil {

		Logger.Error(ErrUnmappingFile.Error())

		return ErrUnmappingFile

	}

	if err := fileMapping.file.Close(); err != nil {

		Logger.Error("error closing file", zap.Error(err))

		return err

	}

	return nil

}

func (fileMapping *FileMapping) WriteAt(data []byte, offset uint64) error {

	fileMapping.lock.Lock()

	defer fileMapping.lock.Unlock()

	// Check if the file size is sufficient
	if int(offset)+len(data) > len(fileMapping.mapping) {

		if err := truncateFile(fileMapping); err != nil {

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

			Logger.Info("error opening new File for: ", zap.Uint32("partitionId:", partitionId), zap.Error(err))

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

	if err := mapping.UnmapFile(); err != nil {

		return err

	}

	delete(pool.pool, partitionId)

	return nil
}

func (pool *OpenFilesPool) Close() {

	pool.lock.Lock()
	defer pool.lock.Unlock()

	for _, fileMapping := range pool.pool {

		if err := syscall.Munmap(fileMapping.mapping); err != nil {

			Logger.Error("error unmapping file", zap.String("fileName", fileMapping.file.Name()), zap.Error(err))

		}

		if err := fileMapping.file.Close(); err != nil {

			Logger.Error("error closing file", zap.String("fileName", fileMapping.file.Name()), zap.Error(err))

		}

	}

}
