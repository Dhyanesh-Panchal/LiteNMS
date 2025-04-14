package main

import (
	. "datastore/containers"
	. "datastore/db"
	. "datastore/reader"
	. "datastore/server"
	. "datastore/utils"
	"log"
)

func main() {

	err := LoadConfig()

	if err != nil {

		log.Println("Error loading config:", err)

		return

	}
	globalShutdown := InitShutdownHandler(4)

	dataWriteChannel := make(chan []PolledDataPoint, DataWriteChannelSize)

	queryChannel := make(chan Query, QueryChannelSize)

	queryResultChannel := make(chan Result, QueryChannelSize)

	go InitDB(dataWriteChannel, queryChannel, queryResultChannel, globalShutdown)

	go InitPollListener(dataWriteChannel, globalShutdown)

	go InitQueryHandler(queryChannel, queryResultChannel, globalShutdown)

	<-globalShutdown

	close(dataWriteChannel)

	close(queryChannel)

}
