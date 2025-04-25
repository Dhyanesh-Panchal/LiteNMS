package server

import (
	"github.com/goccy/go-json"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	. "poller/poller"
	. "poller/utils"
	"sync"
)

func InitSender(pollResultChannel chan PolledDataPoint, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	context, err := zmq.NewContext()

	if err != nil {

		panic(err)

	}

	socket, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		panic(err)

	}

	err = socket.Connect("tcp://localhost:" + PollSenderPort)

	if err != nil {

		panic(err)

	}

	for dataPoint := range pollResultChannel {

		dataBytes, _ := json.Marshal(dataPoint)

		_, err = socket.SendBytes(dataBytes, 0)

		if err != nil {

			Logger.Error("error sending dataPoint ", zap.Any("dataPoint", dataPoint), zap.Error(err))

		}

	}

	Logger.Info("Sender exiting")

}
