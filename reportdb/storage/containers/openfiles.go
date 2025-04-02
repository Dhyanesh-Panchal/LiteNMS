package containers

import (
	"fmt"
	"log"
	"os"
	"path"
	. "reportdb/config"
	. "reportdb/global"
	. "reportdb/utils"
	"strconv"
	"sync"
	"syscall"
)

type OpenFileMapping struct {
	mapping []byte

	file *os.File

	fileInfo os.FileInfo

	lock sync.RWMutex

	// Add functionality of access count to periodically unmap less used files.
}

func NewOpenFileMapping(key FilesPoolKey) (*OpenFileMapping, error) {
	filePath := path.Join(GetStorageDir(key.Date), strconv.Itoa(int(key.CounterId)), strconv.Itoa(int(key.PartitionIndex))+".bin")

	// Ensure that storage directory is present
	err := os.MkdirAll(path.Join(GetStorageDir(key.Date), strconv.Itoa(int(key.CounterId))), 0755)

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0655)
	if err != nil {

		return nil, err

	}

	fileStats, err := file.Stat()

	if err != nil {
		return nil, err
	}

	if fileStats.Size() == 0 {

		// New file created, truncate to initial size
		err := os.Truncate(file.Name(), 4096)
		if err != nil {
			return nil, err
		}

		fileStats, err = file.Stat()

		if err != nil {
			return nil, err
		}

	}

	fileMapping, err := syscall.Mmap(int(file.Fd()), 0, int(fileStats.Size()), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {

		fmt.Println(err)

		return nil, err

	}

	return &OpenFileMapping{

		mapping: fileMapping,

		file: file,

		fileInfo: fileStats,
	}, nil

}

func truncateFile(fileMapping *OpenFileMapping) error {

	newSize := fileMapping.fileInfo.Size() + int64(FileSizeGrowthDelta)

	if err := os.Truncate(fileMapping.file.Name(), newSize); err != nil {
		log.Println("Error truncating file", fileMapping.file.Name(), err)

		return err
	}

	// remap the mapping.

	if err := syscall.Munmap(fileMapping.mapping); err != nil {
		log.Println("Error unmapping file", fileMapping.file.Name(), err)
	}

	newMapping, err := syscall.Mmap(int(fileMapping.file.Fd()), 0, int(newSize), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)

	if err != nil {
		log.Println("Error creating mapping file", fileMapping.file.Name(), err)
		return err
	}

	fileMapping.mapping = newMapping

	fileMapping.fileInfo, err = fileMapping.file.Stat()

	return nil

}

func (fileMapping *OpenFileMapping) WriteAt(data []byte, offset int) error {

	fileMapping.lock.Lock()

	defer fileMapping.lock.Unlock()

	// Check if the file size is sufficient

	if offset+len(data) > int(fileMapping.fileInfo.Size()) {

		err := truncateFile(fileMapping)

		if err != nil {

			return err

		}

	}

	copy(fileMapping.mapping[offset:offset+len(data)], data)

	return nil

}

type FilesPoolKey struct {
	CounterId uint16

	PartitionIndex uint32

	Date Date
}
type OpenFilesPool struct {
	pool map[FilesPoolKey]*OpenFileMapping

	lock sync.Mutex
}

func NewOpenFilesPool() *OpenFilesPool {

	return &OpenFilesPool{pool: make(map[FilesPoolKey]*OpenFileMapping)}

}

func (pool *OpenFilesPool) put(key FilesPoolKey, mapping *OpenFileMapping) {
	pool.lock.Lock()
	defer pool.lock.Unlock()

}

func (pool *OpenFilesPool) Get(key FilesPoolKey) (*OpenFileMapping, error) {

	pool.lock.Lock()

	defer pool.lock.Unlock()

	if mapping, ok := pool.pool[key]; ok {

		return mapping, nil

	} else {
		// File Not present

		// Create new

		mapping, err := NewOpenFileMapping(key)

		if err != nil {

			log.Println("Error opening new File for: ", key, err)

			return nil, err

		}

		pool.pool[key] = mapping

		return mapping, nil
	}
}
