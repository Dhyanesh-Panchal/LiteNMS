package containers

import (
	. "reportdb/config"
	. "reportdb/storage"
	"strconv"
	"sync"
)

type StoragePoolKey struct {
	Date Date

	CounterId uint16
}

type StoragePool struct {
	storagePool map[StoragePoolKey]*Storage
	lock        sync.Mutex
}

func NewOpenStoragePool() *StoragePool {

	return &StoragePool{storagePool: make(map[StoragePoolKey]*Storage)}

}

func (pool *StoragePool) AcquireStorage(key StoragePoolKey) (*Storage, error) {

	pool.lock.Lock()

	defer pool.lock.Unlock()

	if storage, ok := pool.storagePool[key]; ok {

		return storage, nil

	}

	// Storage not in pool. Get new storage.

	storagePath := ProjectRootPath + "/storage/data/" + key.Date.Format() + "/" + strconv.Itoa(int(key.CounterId))

	newStorage, err := NewStorage(storagePath, PartitionCount, BlockSize)

	if err != nil {

		return nil, err

	}

	pool.storagePool[key] = newStorage

	return newStorage, nil

}
