package containers

import (
	. "datastore/utils"
	"github.com/dgraph-io/ristretto"
	"strconv"
)

var DataPointsCache *ristretto.Cache

func InitDataPointsCache() error {

	config := ristretto.Config{
		NumCounters: MaxCacheKeys,
		MaxCost:     MaxCacheSizeInMB * 1024 * 1024,
		BufferItems: 64,
	}

	var err error

	DataPointsCache, err = ristretto.NewCache(&config)

	if err != nil {

		return err

	}

	return nil

}

func CreateCacheKey(storageKey StoragePoolKey, objectId uint32) string {

	return storageKey.Date.Format() + strconv.Itoa(int(storageKey.CounterId)) + strconv.Itoa(int(objectId))

}
