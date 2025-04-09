package storage

import (
	. "datastore/storage/containers"
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

	for len(data) > 0 {

		writableDataBytes := intMin(len(data), int(remainingBlockCapacity))

		writeOffset := objectBlocks[len(objectBlocks)-1].Offset + uint64(index.BlockSize-remainingBlockCapacity)

		err := file.WriteAt(data[:writableDataBytes], writeOffset)

		if err != nil {
			return err
		}

		// Update the Index Metadata
		newBlockCapacity := remainingBlockCapacity - uint32(writableDataBytes)

		index.UpdateObjectBlockCapacity(key, newBlockCapacity)

		//Re-slice for remaining dataPoints
		data = data[writableDataBytes:]

		if len(data) > 0 {
			// Data is remaining hence get new Block

			newBlockOffset := index.GetNextAvailableBlockOffset()

			objectBlocks = index.AppendNewObjectBlock(key,

				ObjectBlock{

					Offset: newBlockOffset,

					RemainingCapacity: index.BlockSize,
				})

			remainingBlockCapacity = index.BlockSize

		} else {

			if writableDataBytes == int(remainingBlockCapacity) {

				// Data was Just Sufficient, hence new Block needs to be assigned

				newBlockOffset := index.GetNextAvailableBlockOffset()

				objectBlocks = index.AppendNewObjectBlock(key,

					ObjectBlock{

						Offset: newBlockOffset,

						RemainingCapacity: index.BlockSize,
					})

			}
		}

	}

	return nil
}

func intMin(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
