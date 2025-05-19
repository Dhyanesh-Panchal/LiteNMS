package main

import (
	. "datastore/containers"
	. "datastore/db"
	. "datastore/query"
	. "datastore/server"
	. "datastore/utils"
	"log"
	"sync"
)

func main() {

	if err := LoadConfig(); err != nil {

		log.Fatal("error loading config:", err)

	}

	if err := InitLogger(); err != nil {

		log.Fatal("error initializing logger", err)

	}

	go InitProfiling()

	globalShutdown := InitShutdownHandler(4)

	var globalShutdownWaitGroup sync.WaitGroup

	dataWriteChannel := make(chan []PolledDataPoint, DataWriteChannelSize)

	queryReceiveChannel := make(chan Query, QueryChannelSize)

	queryResultChannel := make(chan Result, QueryChannelSize)

	globalShutdownWaitGroup.Add(4)

	go InitDB(dataWriteChannel, queryReceiveChannel, queryResultChannel, globalShutdown, &globalShutdownWaitGroup)

	go InitPollListener(dataWriteChannel, globalShutdown, &globalShutdownWaitGroup)

	go InitQueryListener(queryReceiveChannel, globalShutdown, &globalShutdownWaitGroup)

	go InitQueryResultSender(queryResultChannel, &globalShutdownWaitGroup)

	<-globalShutdown

	Logger.Info("closing dataWrite and queryReceive channel")

	close(dataWriteChannel)

	Logger.Info("main waiting for globalShutdownWaitGroup to finish")

	globalShutdownWaitGroup.Wait()

}
