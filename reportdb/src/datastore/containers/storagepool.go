package containers

import (
	. "datastore/storage"
	. "datastore/utils"
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

func (pool *StoragePool) GetStorage(key StoragePoolKey, createIfNotExist bool) (*Storage, error) {

	pool.lock.Lock()

	defer pool.lock.Unlock()

	if storage, ok := pool.storagePool[key]; ok {

		return storage, nil

	}

	// Storage not in pool. Get new storage.

	storagePath := StorageDirectory + "/" + key.Date.Format() + "/" + strconv.Itoa(int(key.CounterId))

	newStorage, err := NewStorage(storagePath, Partitions, BlockSize, createIfNotExist)

	if err != nil {

		return nil, err

	}

	pool.storagePool[key] = newStorage

	return newStorage, nil

}

func (pool *StoragePool) CloseStorage(key StoragePoolKey) {

	pool.lock.Lock()
	defer pool.lock.Unlock()

	if storage, ok := pool.storagePool[key]; ok {

		storage.CloseStorage()

		delete(pool.storagePool, key)

	}

}

func (pool *StoragePool) ClosePool() {

	for _, storage := range pool.storagePool {

		storage.CloseStorage()

	}

	clear(pool.storagePool)

}
