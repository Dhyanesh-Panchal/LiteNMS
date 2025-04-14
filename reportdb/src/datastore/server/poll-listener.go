package server

import (
	. "datastore/containers"
	. "datastore/utils"
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"log"
)

func InitPollListener(dataChannel chan<- []PolledDataPoint, globalShutdown <-chan bool) {

	context, err := zmq.NewContext()

	if err != nil {

		log.Println("Error initializing poll listener context:", err)

		return

	}

	shutDown := make(chan bool)

	go pollListener(context, dataChannel, shutDown)

	// Listen for global shutdown
	<-globalShutdown

	// Send shutdown to socket
	shutDown <- true

	err = context.Term()

	if err != nil {

		log.Println("Error terminating poll listener context:", err)

	}

	// Wait for socket to close.
	<-shutDown

}

func pollListener(context *zmq.Context, dataChannel chan<- []PolledDataPoint, shutDown chan bool) {

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

				log.Println("Error closing poll listener socket ", err)

			}

			// Acknowledge shutDown
			shutDown <- true

			return

		default:

			dataBytes, err := socket.RecvBytes(0)

			if err != nil {

				log.Println("Error receiving poll data", err)

				continue

			}

			var dataPoints []PolledDataPoint

			err = json.Unmarshal(dataBytes, &dataPoints)

			if err != nil {

				log.Println("Error unmarshalling poll data", err)

				continue
			}

			dataChannel <- dataPoints

		}

	}

}
