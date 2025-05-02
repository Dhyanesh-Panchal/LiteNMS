package main

import (
	. "datastore/containers"
	. "datastore/db"
	. "datastore/query"
	. "datastore/server"
	. "datastore/utils"
	"go.uber.org/zap"
	"sync"
)

func main() {

	err := LoadConfig()

	InitLogger()

	go InitProfiling()

	if err != nil {

		Logger.Error("error loading config:", zap.Error(err))

		return

	}
	globalShutdown := InitShutdownHandler(4)

	var globalShutdownWaitGroup sync.WaitGroup

	dataWriteChannel := make(chan []PolledDataPoint, DataWriteChannelSize)

	queryReceiveChannel := make(chan Query, QueryChannelSize)

	queryResultChannel := make(chan Result, QueryChannelSize)

	globalShutdownWaitGroup.Add(4)

	go InitDB(dataWriteChannel, queryReceiveChannel, queryResultChannel, globalShutdown, &globalShutdownWaitGroup)

	go InitPollListener(dataWriteChannel, globalShutdown, &globalShutdownWaitGroup)

	go InitQueryListener(queryReceiveChannel, globalShutdown, &globalShutdownWaitGroup)

	go InitQueryResultPublisher(queryResultChannel, &globalShutdownWaitGroup)

	<-globalShutdown

	Logger.Info("closing dataWrite and queryReceive channel")

	close(dataWriteChannel)

	Logger.Info("main waiting for globalShutdownWaitGroup to finish")

	globalShutdownWaitGroup.Wait()

}
