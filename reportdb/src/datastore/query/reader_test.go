package query

import (
	. "datastore/containers"
	"datastore/utils"
	"fmt"
	"sync"
	"testing"
)

func TestReader(t *testing.T) {

	err := utils.LoadConfig()

	if err != nil {

		t.Error(err)

	}

	readerRequestChannel := make(chan ReaderRequest, 100)

	readerResponseChannel := make(chan ReaderResponse, 10)

	storagePool := InitStoragePool()

	var readersWaitGroup sync.WaitGroup

	readersWaitGroup.Add(1)

	go Reader(readerRequestChannel, readerResponseChannel, storagePool, &readersWaitGroup)

	from := uint32(1747107000)

	to := uint32(1747146600)

	for requestIndex := range 10 {
		request := ReaderRequest{
			RequestIndex: requestIndex,
			StorageKey: StoragePoolKey{
				Date:      UnixToDate(from),
				CounterId: 2,
			},
			From: from,
			To:   to,
			//ObjectIds: []uint32{169093219, 169093224, 2130706433},
			ObjectIds: []uint32{2886731972},
		}

		readerRequestChannel <- request
	}

	for index := range 10 {

		data := <-readerResponseChannel

		fmt.Println("\n\n", index)

		fmt.Println(data)
	}

}
