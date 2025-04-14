package main

import (
	. "datastore/containers"
	. "datastore/db"
	. "datastore/reader"
	. "datastore/server"
	. "datastore/utils"
	"log"
	"sync"
)

func main() {

	err := LoadConfig()

	if err != nil {

		log.Println("Error loading config:", err)

		return

	}
	globalShutdown := InitShutdownHandler(4)

	var globalShutdownWaitGroup sync.WaitGroup

	dataWriteChannel := make(chan []PolledDataPoint, DataWriteChannelSize)

	queryReceiveChannel := make(chan Query, QueryChannelSize)

	queryResultChannel := make(chan Result, QueryChannelSize)

	globalShutdownWaitGroup.Add(3)

	go InitDB(dataWriteChannel, queryReceiveChannel, queryResultChannel, globalShutdown, &globalShutdownWaitGroup)

	go InitPollListener(dataWriteChannel, globalShutdown, &globalShutdownWaitGroup)

	go InitQueryListener(queryReceiveChannel, queryResultChannel, globalShutdown, &globalShutdownWaitGroup)

	go InitProfiling()

	<-globalShutdown

	log.Println("Closing dataWrite and queryReceive channel")

	close(dataWriteChannel)

	close(queryReceiveChannel)

	log.Println("waiting for globalShutdownWaitGroup to finish")

	globalShutdownWaitGroup.Wait()

}
