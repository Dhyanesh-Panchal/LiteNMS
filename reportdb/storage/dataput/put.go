package dataput

import (
	"fmt"
	"math"
	. "reportdb/config"
	. "reportdb/storage/containers"
	"reportdb/utils"
)

func DiskWrite(data []DataPoint, key Key, file *OpenFileMapping, index *Index) error {

	if index.DataType == "string" {

		return writeStringData(data, key, file, index)

	} else {

		return writeNumericData(data, key, file, index)
	}

}

func writeNumericData(data []DataPoint, key Key, file *OpenFileMapping, index *Index) error {

	totalDataPointSize := index.DataPointSize + 4 // 4-bytes for timestamp

	// Index Parameters
	objectIndex := index.GetIndexObjectBlocks(key.ObjectId)

	var remainingBlockCapacity uint32

	if objectIndex != nil {

		fmt.Println("Index Found for ObjectId:", key.ObjectId)

		remainingBlockCapacity = objectIndex[len(objectIndex)-1].RemainingCapacity

	} else {
		// New Object, assign new block

		newBlockOffset := index.GetNextAvailableBlockOffset(key.ObjectId % PartitionCount)

		index.AppendNewObjectBlock(key.ObjectId,

			ObjectBlock{

				Offset: newBlockOffset,

				RemainingCapacity: BlockSize,

				MinTimeStamp: math.MaxUint32,

				MaxTimeStamp: 0,
			})

		objectIndex = index.GetIndexObjectBlocks(key.ObjectId)

		//currentBlockOffsets.SetOffset(key.ObjectId, newBlockOffset)

		remainingBlockCapacity = BlockSize

	}

	for len(data) > 0 {

		newMinTimestamp, newMaxTimestamp := index.GetLastObjectBlockMetadata(key.ObjectId)

		//------------------fmt.Println("Old Metadata: ", newMinTimestamp, newMaxTimestamp)

		writableDataPointsCount := utils.IntMin(len(data), int(remainingBlockCapacity/totalDataPointSize))

		dataBytes := make([]byte, 0, writableDataPointsCount*int(totalDataPointSize))

		for i := 0; i < writableDataPointsCount; i++ {

			// maintain new minmax
			newMinTimestamp = utils.UInt32Min(newMinTimestamp, data[i].Timestamp)
			newMaxTimestamp = utils.UInt32Max(newMaxTimestamp, data[i].Timestamp)

			dataBytes = append(dataBytes, data[i].Serialize(totalDataPointSize)...)

		}

		offset := BlockSize - remainingBlockCapacity

		err := file.WriteAt(dataBytes, int(offset))

		if err != nil {
			return err
		}

		// Update the Index Metadata
		newBlockCapacity := remainingBlockCapacity - uint32(writableDataPointsCount)*totalDataPointSize

		index.UpdateObjectBlockMetadata(key.ObjectId, newBlockCapacity, newMinTimestamp, newMaxTimestamp)

		//Reslice for remaining datapoints
		data = data[writableDataPointsCount:]

		if len(data) > 0 {
			// Data is remaining hence get new Block

			newBlockOffset := index.GetNextAvailableBlockOffset(key.ObjectId % PartitionCount)

			index.AppendNewObjectBlock(key.ObjectId,

				ObjectBlock{

					Offset: newBlockOffset,

					RemainingCapacity: BlockSize,

					MinTimeStamp: math.MaxUint32,

					MaxTimeStamp: 0,
				})

			// currentBlockOffsets.SetOffset(data.ObjectId, newBlockOffset)

			remainingBlockCapacity = BlockSize

		} else {

			if writableDataPointsCount == int(remainingBlockCapacity/totalDataPointSize) {

				// Data was Just Sufficient, hence new Block needs to be assigned

				newBlockOffset := index.GetNextAvailableBlockOffset(key.ObjectId % PartitionCount)

				index.AppendNewObjectBlock(key.ObjectId,

					ObjectBlock{

						Offset: newBlockOffset,

						RemainingCapacity: BlockSize,

						MinTimeStamp: math.MaxUint32,

						MaxTimeStamp: 0,
					})

			}
		}

	}

	// Write Index to disk
	err := index.WriteIndexToFile(key.Date, key.CounterId)

	if err != nil {

		return err

	}

	return nil
}

// TODO: Complete the string writer.
func writeStringData(data []DataPoint, key Key, file *OpenFileMapping, index *Index) error {

	return nil

}
