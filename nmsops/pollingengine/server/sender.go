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

		Logger.Fatal("Could not create sender context", zap.Error(err))

	}

	defer context.Term()

	socket, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		Logger.Fatal("Could not create sender socket", zap.Error(err))

	}

	defer socket.Close()

	err = socket.Connect("tcp://" + BackendHost + ":" + PollSenderPort)

	if err != nil {

		Logger.Fatal("Could not connect sender socket", zap.String("Host", BackendHost), zap.String("Port", PollSenderPort), zap.Error(err))

	}

	dataPointsGroup := make([]PolledDataPoint, 0, PollDataBatchSize)

	size := 0

	for dataPoint := range pollResultChannel {

		dataPointsGroup = append(dataPointsGroup, dataPoint)

		size = (size + 1) % PollDataBatchSize

		if size == 0 {

			dataBytes, _ := json.Marshal(dataPointsGroup)

			_, err = socket.SendBytes(dataBytes, 0)

			if err != nil {

				Logger.Error("error sending dataPointsGroup ", zap.Any("dataPoint", dataPointsGroup), zap.Error(err))

			}

			Logger.Info("Sent dataPointsGroup", zap.Any("dataPoint", dataPointsGroup))

			dataPointsGroup = dataPointsGroup[:0]

		}

	}

	// Send remaining dataPointsGroup
	dataBytes, _ := json.Marshal(dataPointsGroup)

	_, err = socket.SendBytes(dataBytes, 0)

	if err != nil {

		Logger.Error("error sending dataPointsGroup ", zap.Any("dataPoint", dataPointsGroup), zap.Error(err))

	}

	Logger.Info("Sender exiting")

}
