package storage

import (
	. "reportdb/storage/containers"
)

func DiskWrite(key uint32, data []byte, file *FileMapping, index *Index) error {

	var remainingBlockCapacity uint32

	// Get Blocks for the object
	objectBlocks := index.GetIndexObjectBlocks(key)

	if objectBlocks != nil {

		remainingBlockCapacity = objectBlocks[len(objectBlocks)-1].RemainingCapacity

	} else {
		// New Object, assign new block

		newBlockOffset := index.GetNextAvailableBlockOffset()

		objectBlocks = index.AppendNewObjectBlock(key,

			ObjectBlock{

				Offset: newBlockOffset,

				RemainingCapacity: index.BlockSize,
			})

		remainingBlockCapacity = index.BlockSize

	}

	if int(remainingBlockCapacity) < len(data) {

		// not enough space in current block, assign new block

		newBlockOffset := index.GetNextAvailableBlockOffset()

		objectBlocks = index.AppendNewObjectBlock(key,

			ObjectBlock{

				Offset: newBlockOffset,

				RemainingCapacity: index.BlockSize,
			})

		remainingBlockCapacity = index.BlockSize

	}

	writeOffset := objectBlocks[len(objectBlocks)-1].Offset + uint64(index.BlockSize-remainingBlockCapacity)

	err := file.WriteAt(data, writeOffset)

	if err != nil {

		return err

	}

	// Write successfull, update remaining capacity

	index.UpdateObjectBlockCapacity(key, remainingBlockCapacity-uint32(len(data)))

	return nil

}

//func writeNumericData(data []DataPoint, key Key, file *FileMapping, index *Index) error {
//
//	totalDataPointSize := index.DataPointSize + 4 // 4-bytes for timestamp
//
//	// Index Parameters
//	objectIndex := index.GetIndexObjectBlocks(key.ObjectId)
//
//	var remainingBlockCapacity uint32
//
//	if objectIndex != nil {
//
//		fmt.Println("Index Found for ObjectId:", key.ObjectId)
//
//		remainingBlockCapacity = objectIndex[len(objectIndex)-1].RemainingCapacity
//
//	} else {
//		// New Object, assign new block
//
//		newBlockOffset := index.GetNextAvailableBlockOffset(key.ObjectId % PartitionCount)
//
//		index.AppendNewObjectBlock(key.ObjectId,
//
//			ObjectBlock{
//
//				Offset: newBlockOffset,
//
//				RemainingCapacity: BlockSize,
//
//				MinTimeStamp: math.MaxUint32,
//
//				MaxTimeStamp: 0,
//			})
//
//		objectIndex = index.GetIndexObjectBlocks(key.ObjectId)
//
//		//currentBlockOffsets.SetOffset(key.ObjectId, newBlockOffset)
//
//		remainingBlockCapacity = BlockSize
//
//	}
//
//	for len(data) > 0 {
//
//		//------------------fmt.Println("Old Metadata: ", newMinTimestamp, newMaxTimestamp)
//
//		writableData := utils.IntMin(len(data), int(remainingBlockCapacity))
//
//		offset := BlockSize - remainingBlockCapacity
//
//		err := file.WriteAt(data[:writableData], int(offset))
//
//		if err != nil {
//			return err
//		}
//
//		// Update the Index Metadata
//		newBlockCapacity := remainingBlockCapacity - uint32(writableData)
//
//		index.UpdateObjectBlockCapacity(key.ObjectId, newBlockCapacity, newMinTimestamp, newMaxTimestamp)
//
//		//Reslice for remaining datapoints
//		data = data[writableData:]
//
//		if len(data) > 0 {
//			// Data is remaining hence get new Block
//
//			newBlockOffset := index.GetNextAvailableBlockOffset(key.ObjectId % PartitionCount)
//
//			index.AppendNewObjectBlock(key.ObjectId,
//
//				ObjectBlock{
//
//					Offset: newBlockOffset,
//
//					RemainingCapacity: BlockSize,
//
//					MinTimeStamp: math.MaxUint32,
//
//					MaxTimeStamp: 0,
//				})
//
//			// currentBlockOffsets.SetOffset(data.ObjectId, newBlockOffset)
//
//			remainingBlockCapacity = BlockSize
//
//		} else {
//
//			if writableData == int(remainingBlockCapacity/totalDataPointSize) {
//
//				// Data was Just Sufficient, hence new Block needs to be assigned
//
//				newBlockOffset := index.GetNextAvailableBlockOffset(key.ObjectId % PartitionCount)
//
//				index.AppendNewObjectBlock(key.ObjectId,
//
//					ObjectBlock{
//
//						Offset: newBlockOffset,
//
//						RemainingCapacity: BlockSize,
//
//						MinTimeStamp: math.MaxUint32,
//
//						MaxTimeStamp: 0,
//					})
//
//			}
//		}
//
//	}
//
//	// Write Index to disk
//	err := index.WriteIndexToFile(key.Date, key.CounterId)
//
//	if err != nil {
//
//		return err
//
//	}
//
//	return nil
//}
