package main

import (
	"go.uber.org/zap"
	"log"
	. "poller/containers"
	. "poller/poller"
	. "poller/schedular"
	. "poller/server"
	. "poller/utils"
	"sync"
)

const globalShutdownSignalCount = 4

func main() {

	if err := LoadConfig(); err != nil {

		log.Fatal("Error loading config.", err)

	}

	if err := InitLogger(); err != nil {

		log.Fatal("Error initializing logger", err)

	}

	globalShutdownChannel := InitShutdownHandler(globalShutdownSignalCount)

	pollResultChannel := make(chan PolledDataPoint, PollChannelSize)

	pollJobChannel := make(chan PollJob, PollChannelSize)

	deviceList, err := NewDeviceList()

	if err != nil {

		Logger.Error("Error creating device list", zap.Error(err))

		return
	}

	globalShutdownWaitGroup := sync.WaitGroup{}

	globalShutdownWaitGroup.Add(globalShutdownSignalCount - 1)

	// server components
	go InitSender(pollResultChannel, &globalShutdownWaitGroup)

	go InitProvisionListener(deviceList, globalShutdownChannel, &globalShutdownWaitGroup)

	// Pollers
	go InitPollers(pollJobChannel, pollResultChannel, globalShutdownChannel, &globalShutdownWaitGroup)

	// Schedular
	go InitPollScheduler(pollJobChannel, deviceList, globalShutdownChannel, &globalShutdownWaitGroup)

	<-globalShutdownChannel

	Logger.Info("Global shutdown called")

	globalShutdownWaitGroup.Wait()

}
