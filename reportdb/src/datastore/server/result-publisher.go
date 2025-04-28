package server

import (
	. "datastore/query"
	. "datastore/utils"
	"encoding/binary"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	"sync"
)

func InitQueryResultPublisher(queryResultChannel <-chan Result, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	context, err := zmq.NewContext()

	if err != nil {

		Logger.Error("error initializing query result publisher context", zap.Error(err))

		return

	}

	defer context.Term()

	socket, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		Logger.Error("error initializing query result publisher socket", zap.Error(err))

		return

	}

	defer socket.Close()

	err = socket.Bind("tcp://*:" + QueryResultBindPort)

	if err != nil {

		Logger.Error("error binding query result publisher socket", zap.Error(err))

		return

	}

	// Listen for the results

	for result := range queryResultChannel {

		queryId := [8]byte{}

		binary.LittleEndian.PutUint64(queryId[:], result.QueryId)

		resultBytes, err := json.Marshal(result)

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

	Logger.Info("Query result publisher shutting down")

}
