package server

import (
	. "datastore/reader"
	. "datastore/utils"
	"encoding/json"
	"errors"
	zmq "github.com/pebbe/zmq4"
	"log"
	"sync"
)

func InitQueryListener(queryReceiveChannel chan<- Query, queryResultChannel <-chan Result, globalShutdown <-chan bool, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	defer log.Println("Query Listener Exiting")

	context, err := zmq.NewContext()

	if err != nil {

		log.Println("Error initializing query listener context", err)

		return

	}

	shutDown := make(chan bool, 1)

	go queryListener(context, queryReceiveChannel, queryResultChannel, shutDown)

	// Listen for global shutdown
	<-globalShutdown

	// Send shutdown to socket
	shutDown <- true

	err = context.Term()

	if err != nil {

		log.Println("Error terminating query listener context", err)

	}

	// Wait for socket to close.
	<-shutDown

}

func queryListener(context *zmq.Context, queryReceiveChannel chan<- Query, queryResultChannel <-chan Result, shutDown chan bool) {

	socket, err := context.NewSocket(zmq.REP)

	if err != nil {

		log.Fatal("Error initializing query listener socket", err)

	}

	err = socket.Bind("tcp://*:" + QueryListenerBindPort)

	if err != nil {

		log.Fatal("Error binding the ", err)

	}

	for {
		select {

		case <-shutDown:

			err := socket.Close()

			if err != nil {

				log.Println("Error closing query listener socket ", err)

			}

			// Acknowledge shutDown
			shutDown <- true

			return

		default:

			queryBytes, err := socket.RecvBytes(0)

			if err != nil {

				if errors.Is(zmq.AsErrno(err), zmq.ETERM) {

					log.Println("Query Handler ZMQ-Context terminated, closing the socket")

				} else {

					log.Println("Error receiving query ", err)

				}

				continue

			}

			var query Query

			err = json.Unmarshal(queryBytes, &query)

			if err != nil {

				log.Println("Error unmarshalling query ", err)

			}

			// Send it to reader and wait for the response

			queryReceiveChannel <- query

			result := <-queryResultChannel

			log.Println(result)

			resultBytes, err := json.Marshal(result)

			if err != nil {

				log.Println("Error marshalling query result ", err)

			}

			_, err = socket.SendBytes(resultBytes, 0)

			if err != nil {

				log.Println("Error sending query result ", err)

			}

		}

	}

}
