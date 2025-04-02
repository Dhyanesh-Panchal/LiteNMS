package containers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	. "reportdb/config"
	. "reportdb/global"
	. "reportdb/utils"
	"strconv"
	"sync"
	"sync/atomic"
)

type ObjectBlock struct {
	Offset uint64 `json:"offset"`

	RemainingCapacity uint32 `json:"remaining_capacity"`

	MinTimeStamp uint32 `json:"min_time_stamp"`

	MaxTimeStamp uint32 `json:"max_time_stamp"`
}

type PartitionIndex struct {
	NextFreeBlockOffset uint64 `json:"next_free_block_offset"`

	ObjectIndex map[uint32][]ObjectBlock `json:"object_index"`
}

type Index struct {
	DataType string `json:"data_type"`

	DataPointSize uint32 `json:"data_point_size"`

	PartitionIndex []PartitionIndex `json:"partition_index"`

	mu sync.RWMutex
}

func newIndex(dataType string, dataSize uint32) *Index {

	partitionIndices := make([]PartitionIndex, PartitionCount)

	for index := range PartitionCount {

		partitionIndices[index] = PartitionIndex{

			ObjectIndex: make(map[uint32][]ObjectBlock),
		}

	}

	return &Index{

		PartitionIndex: partitionIndices,

		DataType: dataType,

		DataPointSize: dataSize,
	}

}

func loadIndex(date Date, counterId uint16) (*Index, error) {

	fmt.Println("New Index file demanded for ", date, counterId)

	indexFilePath := path.Join(GetStorageDir(date), strconv.Itoa(int(counterId)), "index.json")

	indexBytes, err := os.ReadFile(indexFilePath)

	if err != nil {

		if os.IsNotExist(err) {

			// ensure storage directory is present

			err := os.MkdirAll(path.Join(GetStorageDir(date), strconv.Itoa(int(counterId))), 0755)

			file, err := os.Create(indexFilePath)

			if err != nil {

				return nil, err

			} else {

				defer func() {
					if err := file.Close(); err != nil {

						log.Println("Failed to close index file")

					}
				}()

				// Create New Index

				index := newIndex(CounterConfig[counterId]["dataType"].(string), CounterConfig[counterId]["dataSize"].(uint32))

				indexBytes, err := json.MarshalIndent(index, "", "  ")

				if err != nil {
					return nil, err
				}

				_, err = file.Write(indexBytes)

				if err != nil {
					return nil, err
				}

				return index, nil
			}

		}

		return nil, err
	}

	var index Index

	err = json.Unmarshal(indexBytes, &index)

	if err != nil {

		return nil, err

	}

	return &index, nil
}

func (index *Index) GetIndexObjectBlocks(objectId uint32) []ObjectBlock {

	index.mu.RLock()

	defer index.mu.RUnlock()

	if objectBlocks, ok := index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId]; !ok {

		return nil

	} else {

		return objectBlocks

	}
}

func (index *Index) AppendNewObjectBlock(objectId uint32, objectBlock ObjectBlock) {

	index.mu.Lock()

	defer index.mu.Unlock()

	index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId] = append(index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId], objectBlock)

}

func (index *Index) GetLastObjectBlockMetadata(objectId uint32) (uint32, uint32) {

	index.mu.RLock()

	defer index.mu.RUnlock()

	lastIndex := len(index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId]) - 1

	return index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId][lastIndex].MinTimeStamp, index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId][lastIndex].MaxTimeStamp

}

func (index *Index) UpdateObjectBlockMetadata(objectId uint32, newBlockCapacity uint32, newMinTimestamp uint32, newMaxTimestamp uint32) {

	index.mu.Lock()

	defer index.mu.Unlock()

	lastIndx := len(index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId]) - 1

	index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId][lastIndx].RemainingCapacity = newBlockCapacity

	index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId][lastIndx].MinTimeStamp = newMinTimestamp

	index.PartitionIndex[objectId%PartitionCount].ObjectIndex[objectId][lastIndx].MaxTimeStamp = newMaxTimestamp

}

func (index *Index) WriteIndexToFile(date Date, counterId uint16) error {
	// TODO: Change from MarshalIndent to Only Marshal
	index.mu.RLock()
	defer index.mu.RUnlock()

	indexBytes, err := json.MarshalIndent(index, "", "  ")

	indexFilePath := path.Join(GetStorageDir(date), strconv.Itoa(int(counterId)), "index.json")

	indexFile, err := os.Create(indexFilePath)

	if err != nil {
		return err
	}

	defer func() {
		if err := indexFile.Close(); err != nil {
			log.Println("Failed to close index file")
		}
	}()

	if _, err = indexFile.Write(indexBytes); err != nil {
		return err
	}

	return nil
}

func (index *Index) GetNextAvailableBlockOffset(partitionId uint32) uint64 {

	nextOffset := atomic.SwapUint64(&index.PartitionIndex[partitionId].NextFreeBlockOffset, index.PartitionIndex[partitionId].NextFreeBlockOffset+uint64(BlockSize))

	return nextOffset

}

type IndexPoolKey struct {
	Date Date

	CounterId uint16
}

type IndexPool struct {
	pool map[IndexPoolKey]*Index

	lock sync.Mutex
}

func NewIndexPool() *IndexPool {

	return &IndexPool{

		pool: make(map[IndexPoolKey]*Index),
	}

}

func (indexPool *IndexPool) Get(key IndexPoolKey) (*Index, error) {

	indexPool.lock.Lock()

	defer indexPool.lock.Unlock()

	if index, ok := indexPool.pool[key]; ok {

		return index, nil

	} else {

		// load the corresponding index file

		index, err := loadIndex(key.Date, key.CounterId)

		if err != nil {

			log.Println("Error opening new index for: ", key, err)

			return nil, err

		}

		// Update the Pool
		indexPool.pool[key] = index

		return index, nil

	}

}
