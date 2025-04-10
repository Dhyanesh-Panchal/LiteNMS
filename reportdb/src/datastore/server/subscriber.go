package server

import (
	. "datastore/containers"
	. "datastore/utils"
	"encoding/json"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
)

type Subscriber struct {
	context *zmq.Context
}

func InitPollSubscriber(dataChannel chan<- []PolledDataPoint) {

	context, err := zmq.NewContext()

	if err != nil {

		log.Println("Error initializing poll subscriber zmq-context", err)

		return

	}

	shutDown := make(chan bool)

	go Listener(context, dataChannel, shutDown)

	// Listen for global shutdown
	<-GlobalShutdown

	// Send shutdown to socket
	shutDown <- true

	err = context.Term()

	if err != nil {

		log.Println("Error terminating poll subscriber", err)

	}

	// Wait for socket to close.
	<-shutDown

	return

}

func Listener(context *zmq.Context, dataChannel chan<- []PolledDataPoint, shutDown chan bool) {

	socket, err := context.NewSocket(zmq.SUB)

	if err != nil {

		log.Fatal("Error initializing poll subscriber zmq-socket", err)

	}

	err = socket.Bind("tcp://*:" + SubscriberBindPort)
	socket.SetSubscribe("")
	if err != nil {

		log.Fatal("Error binding the subscriber", err)

	}

	for {
		select {
		case <-shutDown:
			log.Println("Shutting down poll subscriber")

			err := socket.Close()

			if err != nil {

				log.Println("Error closing poll subscriber socket ", err)

			}

			// Acknowledge shutDown
			shutDown <- true

			return

		default:

			dataBytes, err := socket.RecvBytes(0)

			if err != nil {

				log.Println("Error receiving data", err)

				continue

			}

			var dataPoints []PolledDataPoint

			err = json.Unmarshal(dataBytes, &dataPoints)

			if err != nil {

				log.Println("Error unmarshalling data", err)

				continue
			}

			fmt.Println("Received data:", dataPoints)

			dataChannel <- dataPoints

		}

	}

}
