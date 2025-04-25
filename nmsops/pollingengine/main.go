package main

import (
	"context"
	"go.uber.org/zap"
	. "poller/containers"
	. "poller/poller"
	. "poller/schedular"
	. "poller/server"
	. "poller/utils"
	"sync"
)

func main() {
	err := LoadConfig()

	if err != nil {

		Logger.Error("Error loading config.", zap.Error(err))

	}

	globalShutdownChannel := InitShutdownHandler(3)

	globalContext, globalContextCancel := context.WithCancel(context.Background())

	pollResultChannel := make(chan PolledDataPoint, PollChannelSize)

	pollJobChannel := make(chan PollJob, PollChannelSize)

	deviceList := NewDeviceList(globalContext)

	globalShutdownWaitGroup := sync.WaitGroup{}

	// server components

	globalShutdownWaitGroup.Add(2)

	go InitSender(pollResultChannel, &globalShutdownWaitGroup)

	go InitProvisionListener(deviceList, globalShutdownChannel, &globalShutdownWaitGroup)

	// Pollers

	globalShutdownWaitGroup.Add(PollWorkers)

	for range PollWorkers {

		go Poller(pollJobChannel, pollResultChannel, &globalShutdownWaitGroup)

	}

	// Schedular

	go InitPollScheduler(pollJobChannel, deviceList, globalShutdownChannel, &globalShutdownWaitGroup)

	<-globalShutdownChannel

	Logger.Info("Global shutdown called")

	close(pollJobChannel)

	close(pollResultChannel)

	globalContextCancel()

	globalShutdownWaitGroup.Wait()

}
