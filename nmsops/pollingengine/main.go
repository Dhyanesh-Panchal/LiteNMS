package main

import (
	"go.uber.org/zap"
	. "poller/containers"
	. "poller/poller"
	. "poller/schedular"
	. "poller/server"
	. "poller/utils"
	"sync"
)

const globalShutdownSignalCount = 4

func main() {
	err := LoadConfig()

	if err != nil {

		Logger.Error("Error loading config.", zap.Error(err))

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

	// server components

	globalShutdownWaitGroup.Add(globalShutdownSignalCount - 1)

	go InitSender(pollResultChannel, &globalShutdownWaitGroup)

	go InitProvisionListener(deviceList, globalShutdownChannel, &globalShutdownWaitGroup)

	go InitPollers(pollJobChannel, pollResultChannel, globalShutdownChannel, &globalShutdownWaitGroup)

	// Schedular

	go InitPollScheduler(pollJobChannel, deviceList, globalShutdownChannel, &globalShutdownWaitGroup)

	<-globalShutdownChannel

	Logger.Info("Global shutdown called")

	globalShutdownWaitGroup.Wait()

}
