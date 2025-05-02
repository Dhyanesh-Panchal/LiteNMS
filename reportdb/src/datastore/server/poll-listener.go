package server

import (
	. "datastore/containers"
	. "datastore/utils"
	"errors"
	zmq "github.com/pebbe/zmq4"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"log"
	"sync"
)

func InitPollListener(dataChannel chan<- []PolledDataPoint, globalShutdown <-chan bool, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	defer Logger.Info("Poll Listener Exiting")

	context, err := zmq.NewContext()

	if err != nil {

		Logger.Error("error initializing poll listener context:" + err.Error())

		return

	}

	shutDown := make(chan bool, 1)

	go pollListener(context, dataChannel, shutDown)

	// Listen for global shutdown
	<-globalShutdown

	// Send shutdown to socket
	shutDown <- true

	err = context.Term()

	if err != nil {

		Logger.Error("error terminating poll listener context:", zap.Error(err))

	}

	// Wait for socket to close.
	<-shutDown

}

func pollListener(context *zmq.Context, dataWriteChannel chan<- []PolledDataPoint, shutDown chan bool) {

	socket, err := context.NewSocket(zmq.PULL)

	if err != nil {

		log.Fatal("Error initializing poll listener socket:", err)

	}

	err = socket.Bind("tcp://*:" + PollListenerBindPort)

	if err != nil {

		log.Fatal("Error binding the poll listener socket: ", err)

	}

	for {

		select {

		case <-shutDown:

			err := socket.Close()

			if err != nil {

				Logger.Error("error closing poll listener socket ", zap.Error(err))

			}

			// Acknowledge shutDown
			shutDown <- true

			return

		default:

			dataBytes, err := socket.RecvBytes(0)

			if err != nil {

				if errors.Is(zmq.AsErrno(err), zmq.ETERM) {

					Logger.Info("Poll listener ZMQ-Context terminated, closing the socket")

				} else {

					Logger.Error("error receiving poll data", zap.Error(err))

				}

				continue

			}

			var dataPoints []PolledDataPoint

			if err := msgpack.Unmarshal(dataBytes, &dataPoints); err != nil {

				Logger.Error("error unmarshalling poll data", zap.Error(err))

				continue
			}

			dataWriteChannel <- dataPoints

		}

	}

}
