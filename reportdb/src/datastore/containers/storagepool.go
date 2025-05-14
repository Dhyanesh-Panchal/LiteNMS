package containers

import (
	. "datastore/storage"
	. "datastore/utils"
	"go.uber.org/zap"
	"strconv"
	"sync"
	"time"
)

type StoragePoolKey struct {
	Date Date

	CounterId uint16
}

type StoragePool struct {
	pool map[StoragePoolKey]*Storage

	accessCount map[StoragePoolKey]int

	cleanupTicker *time.Ticker

	lock sync.Mutex
}

func InitStoragePool() *StoragePool {

	storagePool := &StoragePool{
		pool: make(map[StoragePoolKey]*Storage),

		accessCount: make(map[StoragePoolKey]int),

		cleanupTicker: time.NewTicker(time.Second * time.Duration(StorageCleanupInterval)),
	}

	go storagePoolCleanup(storagePool)

	return storagePool

}

func (storagePool *StoragePool) GetStorage(key StoragePoolKey, createIfNotExist bool) (*Storage, error) {

	storagePool.lock.Lock()

	defer storagePool.lock.Unlock()

	if storage, ok := storagePool.pool[key]; ok {

		storagePool.accessCount[key]++

		return storage, nil

	}

	// Storage not in pool. Get new storage.

	// First clear the storage

	storagePath := StorageDirectory + "/" + key.Date.Format() + "/" + strconv.Itoa(int(key.CounterId))

	newStorage, err := NewStorage(storagePath, Partitions, BlockSize, createIfNotExist)

	if err != nil {

		return nil, err

	}

	storagePool.pool[key] = newStorage

	storagePool.accessCount[key]++

	Logger.Info("Loaded new storage in pool", zap.Any("Key", key))

	return newStorage, nil

}

func (storagePool *StoragePool) CleanPool() {

	storagePool.lock.Lock()

	defer storagePool.lock.Unlock()

	for key, storage := range storagePool.pool {

		if storagePool.accessCount[key] < 10 {

			storage.ClearStorage()

			delete(storagePool.accessCount, key)

			delete(storagePool.pool, key)

			Logger.Info("Closed storage", zap.Any("Key", key))

		} else {

			storagePool.accessCount[key] = 0

		}

	}

}

func (storagePool *StoragePool) ClosePool() {

	for _, storage := range storagePool.pool {

		storage.ClearStorage()

	}

	clear(storagePool.pool)

}

func storagePoolCleanup(storagePool *StoragePool) {

	for {

		select {

		//case <-shutdownChannel:
		//
		//	storagePool.ClosePool()
		//
		//	return

		case <-storagePool.cleanupTicker.C:
			storagePool.CleanPool()
		}
	}

}
