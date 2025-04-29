package server

import (
	. "datastore/query"
	. "datastore/utils"
	"encoding/json"
	"errors"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	"sync"
)

func InitQueryListener(queryReceiveChannel chan<- Query, globalShutdown <-chan bool, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	context, err := zmq.NewContext()

	if err != nil {

		Logger.Error("error initializing query listener context", zap.Error(err))

		return

	}

	queryListenerShutdown := make(chan struct{}, 1)

	go queryListener(context, queryReceiveChannel, queryListenerShutdown)

	// Listen for global shutdown
	<-globalShutdown

	// Send shutdown to socket
	queryListenerShutdown <- struct{}{}

	err = context.Term()

	if err != nil {

		Logger.Error("error terminating query listener context", zap.Error(err))

	}

	// Wait for socket to close.
	<-queryListenerShutdown

	close(queryReceiveChannel)

}

func queryListener(context *zmq.Context, queryReceiveChannel chan<- Query, queryListenerShutdown chan struct{}) {

	socket, err := context.NewSocket(zmq.PULL)

	if err != nil {

		Logger.Error("Error initializing query listener socket", zap.Error(err))

		return

	}

	err = socket.Bind("tcp://*:" + QueryListenerBindPort)

	if err != nil {

		Logger.Error("Error binding query listener socket", zap.String("port", QueryListenerBindPort), zap.Error(err))

	}

	for {
		select {

		case <-queryListenerShutdown:

			err := socket.Close()

			if err != nil {

				Logger.Error("error closing query listener socket ", zap.Error(err))

			}

			// Acknowledge
			queryListenerShutdown <- struct{}{}

			return

		default:

			queryBytes, err := socket.RecvBytes(0)

			if err != nil {

				if errors.Is(zmq.AsErrno(err), zmq.ETERM) {

					Logger.Info("Query Handler's ZMQ-Context terminated, closing the socket")

				} else {

					Logger.Error("error receiving query ", zap.Error(err))

				}

				continue

			}

			var query Query

			if err = json.Unmarshal(queryBytes, &query); err != nil {

				Logger.Error("error unmarshalling query ", zap.Error(err))

			}

			Logger.Debug("Received query ", zap.Any("query", query))

			queryReceiveChannel <- query

		}

	}

}
