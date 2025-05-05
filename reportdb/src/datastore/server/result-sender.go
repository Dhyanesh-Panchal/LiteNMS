package server

import (
	. "datastore/query"
	. "datastore/utils"
	"encoding/binary"
	zmq "github.com/pebbe/zmq4"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"sync"
)

func InitQueryResultSender(queryResultChannel <-chan Result, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	context, err := zmq.NewContext()

	if err != nil {

		Logger.Error("error initializing query result sender context", zap.Error(err))

		return

	}

	defer func(context *zmq.Context) {

		err := context.Term()

		if err != nil {

			Logger.Error("Error terminating the result sender zmq context")

		}

	}(context)

	socket, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		Logger.Error("error initializing query result sender socket", zap.Error(err))

		return

	}

	defer func(socket *zmq.Socket) {
		err := socket.Close()
		if err != nil {
			Logger.Error("Error terminating the result sender zmq socket")
		}
	}(socket)

	err = socket.Bind("tcp://*:" + QueryResultBindPort)

	if err != nil {

		Logger.Error("error binding query result sender socket", zap.Error(err))

		return

	}

	// Listen for the results

	for result := range queryResultChannel {

		queryId := [8]byte{}

		binary.LittleEndian.PutUint64(queryId[:], result.QueryId)

		resultBytes, err := msgpack.Marshal(result)

		if err != nil {

			Logger.Error("error marshalling query result ", zap.Error(err))

			continue

		}

		message := append(queryId[:], resultBytes...)

		_, err = socket.SendBytes(message, 0)

		if err != nil {

			Logger.Error("error sending query result ", zap.Error(err))

		}

	}

	Logger.Info("Query result sender shutting down")

}
