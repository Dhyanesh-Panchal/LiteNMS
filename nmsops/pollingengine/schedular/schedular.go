package schedular

import (
	"context"
	. "poller/containers"
	. "poller/poller"
	. "poller/utils"
	"sync"
	"time"
)

func InitPollScheduler(pollJobChannel chan<- PollJob, deviceList *DeviceList, globalShutdownChannel <-chan struct{}, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	schedularContext, cancel := context.WithCancel(context.Background())

	var schedularWaitGroup sync.WaitGroup

	schedularWaitGroup.Add(1)

	go scheduler(pollJobChannel, deviceList, schedularContext, &schedularWaitGroup)

	<-globalShutdownChannel

	cancel()

	schedularWaitGroup.Wait()
	
	Logger.Debug("Scheduler Exiting")
}

func scheduler(pollJobChannel chan<- PollJob, deviceList *DeviceList, schedularContext context.Context, schedularWaitGroup *sync.WaitGroup) {

	defer schedularWaitGroup.Done()

	// initialize poll Intervals for the counter

	counterPollIntervals := map[uint16]uint32{}

	for counterId, _ := range CounterConfig {

		counterPollIntervals[counterId] = uint32(CounterConfig[counterId]["pollingInterval"].(float64))

	}

	pollTicker := time.NewTicker(time.Second)

	for {

		select {
		case tick := <-pollTicker.C:

			devicesCredential := deviceList.GetDevices()

			timestamp := uint32(tick.UTC().Unix())

			var qualifiedCounterIds []uint16

			// determine qualified counterIds for corresponding tick
			for counterId, _ := range CounterConfig {

				if timestamp%counterPollIntervals[counterId] == 0 {

					qualifiedCounterIds = append(qualifiedCounterIds, counterId)

				}

			}

			for deviceId, config := range devicesCredential {

				pollJobChannel <- PollJob{
					Timestamp:  timestamp,
					DeviceIP:   deviceId,
					Hostname:   config[0],
					Password:   config[1],
					Port:       config[2],
					CounterIds: qualifiedCounterIds,
				}

			}

		case <-schedularContext.Done():
			pollTicker.Stop()

			Logger.Info("Shutting down scheduler")

			return
		}

	}

}
