package containers

import (
	. "datastore/utils"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

type ObjectBlock struct {
	Offset uint64 `msgpack:"offset" json:"offset"`

	RemainingCapacity uint32 `msgpack:"remaining_capacity" json:"remaining_capacity"`
}

type Index struct {
	BlockSize uint32 `msgpack:"block_size" json:"block_size"`

	NextFreeBlockOffset uint64 `msgpack:"next_free_block_offset" json:"next_free_block_offset"`

	ObjectIndex map[uint32][]ObjectBlock `msgpack:"object_index"  json:"object_index"`

	mu sync.RWMutex
}

func NewIndex(blockSize uint32) *Index {

	return &Index{

		BlockSize: blockSize,

		NextFreeBlockOffset: 0,

		ObjectIndex: make(map[uint32][]ObjectBlock),
	}

}

func loadIndex(partitionId uint32, storagePath string) (*Index, error) {

	indexFilePath := storagePath + "/index_" + strconv.Itoa(int(partitionId)) + ".bin"

	indexBytes, err := os.ReadFile(indexFilePath)

	if err != nil {

		log.Printf("Error reading index file: %v", err)

		return nil, err

	}

	var index Index

	if err = msgpack.Unmarshal(indexBytes, &index); err != nil {

		log.Printf("Error unmarshalling index file: %v", err)

		return nil, err

	}

	return &index, nil
}

func (index *Index) GetIndexObjectBlocks(objectId uint32) []ObjectBlock {

	index.mu.RLock()

	defer index.mu.RUnlock()

	if objectBlocks, ok := index.ObjectIndex[objectId]; !ok {

		return nil

	} else {

		return objectBlocks

	}
}

func (index *Index) AppendNewObjectBlock(objectId uint32, objectBlock ObjectBlock) []ObjectBlock {

	index.mu.Lock()

	defer index.mu.Unlock()

	index.ObjectIndex[objectId] = append(index.ObjectIndex[objectId], objectBlock)

	return index.ObjectIndex[objectId]

}

func (index *Index) GetLastObjectBlockCapacity(objectId uint32) uint32 {

	index.mu.RLock()

	defer index.mu.RUnlock()

	lastIndex := len(index.ObjectIndex[objectId]) - 1

	return index.ObjectIndex[objectId][lastIndex].RemainingCapacity

}

func (index *Index) UpdateObjectBlockCapacity(objectId uint32, newBlockCapacity uint32) {

	index.mu.Lock()

	defer index.mu.Unlock()

	lastIndex := len(index.ObjectIndex[objectId]) - 1

	index.ObjectIndex[objectId][lastIndex].RemainingCapacity = newBlockCapacity

}

func (index *Index) WriteIndexToFile(storagePath string, partitionId uint32) error {

	index.mu.Lock()

	defer index.mu.Unlock()

	indexBytes, err := msgpack.Marshal(index)
	
	indexFilePath := storagePath + "/index_" + strconv.Itoa(int(partitionId)) + ".bin"

	err = os.WriteFile(indexFilePath, indexBytes, 0644)

	if err != nil {

		return err

	}

	return nil
}

func (index *Index) GetNextAvailableBlockOffset() uint64 {

	nextOffset := atomic.SwapUint64(&index.NextFreeBlockOffset, index.NextFreeBlockOffset+uint64(index.BlockSize))

	return nextOffset

}

type IndexPool struct {
	pool map[uint32]*Index

	lock sync.Mutex
}

func NewIndexPool() *IndexPool {

	return &IndexPool{

		pool: make(map[uint32]*Index),
	}

}

func (indexPool *IndexPool) Get(partitionId uint32, storagePath string) (*Index, error) {

	indexPool.lock.Lock()

	defer indexPool.lock.Unlock()

	if index, ok := indexPool.pool[partitionId]; ok {

		return index, nil

	} else {

		// load the corresponding index file

		index, err := loadIndex(partitionId, storagePath)

		if err != nil {

			Logger.Error("error opening new index for: ", zap.String("storagePath", storagePath), zap.Uint32("partitionId", partitionId), zap.Error(err))

			return nil, err

		}

		// Update the Pool
		indexPool.pool[partitionId] = index

		return index, nil

	}

}

func (indexPool *IndexPool) Close(storagePath string) {

	indexPool.lock.Lock()
	defer indexPool.lock.Unlock()

	// Sync changes if any
	for partitionId, index := range indexPool.pool {

		err := index.WriteIndexToFile(storagePath, partitionId)

		if err != nil {

			Logger.Error("error closing index for: ", zap.String("storagePath", storagePath), zap.Uint32("partitionId", partitionId), zap.Error(err))

		}
	}

}
