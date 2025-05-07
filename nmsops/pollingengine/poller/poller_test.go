package poller

import (
	"poller/utils"
	"sync"
	"testing"
)

func TestPoller(t *testing.T) {

	err := utils.LoadConfig()

	if err != nil {

		t.Error(err)

	}

	pollResultChannel := make(chan PolledDataPoint, 10)

	pollJobChannel := make(chan PollJob, 10)

	wg := sync.WaitGroup{}

	shutdownChannel := make(chan struct{}, 1)

	go InitPollers(pollJobChannel, pollResultChannel, shutdownChannel, &wg)

	deviceIp := "172.16.8.71"
	port := "22"

	pollJob := PollJob{
		Timestamp: 1687000000,

		DeviceIP: deviceIp,

		Hostname: "motadata",

		Password: "motadata",

		Port: port,

		CounterIds: []uint16{1, 2, 3},
	}

	for range 10 {

		pollJobChannel <- pollJob
		
	}

	resp := <-pollResultChannel

	t.Log(resp)

	resp = <-pollResultChannel

	t.Log(resp)

	resp = <-pollResultChannel

	t.Log(resp)

	shutdownChannel <- struct{}{}

	wg.Wait()
}
