package schedular

import (
	"context"
	"go.uber.org/zap"
	. "poller/containers"
	. "poller/poller"
	. "poller/utils"
	"sync"
	"time"
)

func InitPollScheduler(pollJobChannel chan<- PollJob, deviceList *DeviceList, globalShutdownChannel <-chan struct{}, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	schedularContext, cancel := context.WithCancel(context.Background())

	for counterId, _ := range CounterConfig {

		go counterSchedule(counterId, pollJobChannel, deviceList, schedularContext)

	}

	<-globalShutdownChannel

	cancel()

}

func counterSchedule(counterId uint16, pollJobChannel chan<- PollJob, deviceList *DeviceList, schedularContext context.Context) {

	pollTicker := time.NewTicker(time.Duration(CounterConfig[counterId]["pollingInterval"].(float64)) * time.Second)

	for {

		select {
		case tick := <-pollTicker.C:

			devicesConfig, devicesPort := deviceList.GetDevices()

			timestamp := uint32(tick.UTC().Unix())

			for deviceId, _ := range devicesConfig {

				pollJobChannel <- PollJob{

					Timestamp:    timestamp,
					DeviceIP:     deviceId,
					DeviceConfig: devicesConfig[deviceId],
					DevicePort:   devicesPort[deviceId],
					CounterId:    counterId,
				}

			}

		case <-schedularContext.Done():
			pollTicker.Stop()

			Logger.Info("Shutting down counter scheduler", zap.Uint16("counterId", counterId))

			return
		}
	}

}
