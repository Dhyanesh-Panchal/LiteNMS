package query

import (
	. "datastore/containers"
	"datastore/utils"
	"fmt"
	"sync"
	"testing"
)

func TestReader(t *testing.T) {

	utils.LoadConfig()

	readerRequestChannel := make(chan ReaderRequest, 100)

	parserWaitChannels := make([]chan map[string]interface{}, 1)

	parserWaitChannels[0] = make(chan map[string]interface{}, 10)

	storagePool := NewOpenStoragePool()

	var readersWaitGroup sync.WaitGroup

	readersWaitGroup.Add(1)

	go Reader(readerRequestChannel, parserWaitChannels, storagePool, &readersWaitGroup)

	from := uint32(1744781400)

	to := uint32(1744805400)

	for requestIndex := range 10 {
		request := ReaderRequest{
			RequestIndex: requestIndex,
			StorageKey: StoragePoolKey{
				Date:      UnixToDate(from),
				CounterId: 2,
			},
			From:      from,
			To:        to,
			ObjectIds: []uint32{169093219, 169093224, 2130706433},
		}

		readerRequestChannel <- request
	}

	for index := range 10 {

		data := <-parserWaitChannels[0]

		fmt.Println("\n\n", index)

		fmt.Println(data)
	}

}
