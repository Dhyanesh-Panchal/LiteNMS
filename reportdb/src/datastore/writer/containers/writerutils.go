package containers

import (
	. "datastore/containers"
)

type WritableObjectData struct {
	StorageKey StoragePoolKey
	ObjectId   uint32
	Values     []DataPoint
}
