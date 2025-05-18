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

	defer func(context *zmq.Context) {

		err := context.Term()

		if err != nil {

			Logger.Error("Error terminating the sender zmq context")

		}

	}(context)

	socket, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		Logger.Fatal("Could not create sender socket", zap.Error(err))

	}

	defer func(socket *zmq.Socket) {

		if err := socket.Close(); err != nil {

			Logger.Error("Error terminating the sender zmq socket")

		}

	}(socket)

	// Set linger to 0 to avoid blocking on close
	if err = socket.SetLinger(0); err != nil {

		Logger.Error("Failed to set linger", zap.Error(err))

	}

	if err = socket.Connect("tcp://" + BackendHost + ":" + PollSenderPort); err != nil {

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

				Logger.Error("error sending dataPoints", zap.Any("dataPoint", dataPointsGroup), zap.Error(err))

			}

			Logger.Info("Sent dataPoints", zap.Any("dataPoint", dataPointsGroup))

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
