package schedular

import (
	. "poller/containers"
	. "poller/utils"
	"sync"
	"time"
)

func InitPollScheduler(pollJobChannel chan<- PollJob, deviceList *DeviceList, globalShutdownChannel <-chan struct{}, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	schedularShutdownChannel := make(chan struct{}, 1)

	go scheduler(pollJobChannel, deviceList, schedularShutdownChannel)

	<-globalShutdownChannel

	schedularShutdownChannel <- struct{}{}

	// Wait for scheduler to exit
	<-schedularShutdownChannel

	close(pollJobChannel)

	Logger.Debug("Scheduler Exiting")
}

func scheduler(pollJobChannel chan<- PollJob, deviceList *DeviceList, schedularShutdownChannel chan struct{}) {

	// initialize poll Intervals for the counter
	counterPollIntervals := map[uint16]uint32{}

	var baseTickInterval uint32

	for counterId, _ := range CounterConfig {

		counterPollIntervals[counterId] = uint32(CounterConfig[counterId]["pollingInterval"].(float64))

		if baseTickInterval == 0 {
			// initial baseTick
			baseTickInterval = counterPollIntervals[counterId]

		} else {

			baseTickInterval = gcd(baseTickInterval, counterPollIntervals[counterId])

		}

	}

	pollTicker := time.NewTicker(time.Second * time.Duration(baseTickInterval))

	for {

		select {
		case tick := <-pollTicker.C:

			timestamp := uint32(tick.UTC().Unix())

			var qualifiedCounterIds []uint16

			// determine qualified counterIds for corresponding tick
			for counterId, _ := range CounterConfig {

				if timestamp%counterPollIntervals[counterId] == 0 {

					qualifiedCounterIds = append(qualifiedCounterIds, counterId)

				}

			}

			if len(qualifiedCounterIds) == 0 {

				continue

			}

			jobs := deviceList.PreparePollJobs(timestamp, qualifiedCounterIds)

			for _, job := range jobs {

				pollJobChannel <- job

			}

		case <-schedularShutdownChannel:
			pollTicker.Stop()

			// Acknowledge shutdown
			schedularShutdownChannel <- struct{}{}

			return
		}

	}

}

// Helper function to find GCD of two numbers using Euclidean algorithm
func gcd(a, b uint32) uint32 {

	for b != 0 {

		a, b = b, a%b

	}

	return a

}
