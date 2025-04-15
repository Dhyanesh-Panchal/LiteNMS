package server

import (
	. "datastore/reader"
	. "datastore/utils"
	"encoding/binary"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"log"
	"sync"
)

func InitQueryResultPublisher(queryResultChannel <-chan Result, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	context, err := zmq.NewContext()

	if err != nil {

		log.Println("Error initializing query result publisher context", err)

		return

	}

	defer context.Term()

	socket, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		log.Println("Error initializing query result publisher socket", err)

	}

	defer socket.Close()

	err = socket.Bind("tcp://*:" + QueryResultBindPort)

	if err != nil {

		log.Println("Error binding query result publisher socket", err)

	}

	for result := range queryResultChannel {

		queryId := [8]byte{}

		binary.LittleEndian.PutUint64(queryId[:], result.QueryId)

		resultBytes, err := json.Marshal(result)

		if err != nil {

			log.Println("Error marshalling query result ", err)

			continue

		}

		message := append(queryId[:], resultBytes...)

		_, err = socket.SendBytes(message, 0)

		if err != nil {

			log.Println("Error sending query result ", err)

		}
	}

	log.Println("Query result publisher shutting down")

}
