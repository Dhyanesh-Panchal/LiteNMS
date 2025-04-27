package services

import (
	"encoding/binary"
	zmq "github.com/pebbe/zmq4"
	"log"
)

type ProvisioningPublisher struct {
	context *zmq.Context
	socket  *zmq.Socket
}

func InitProvisioningPublisher() (*ProvisioningPublisher, error) {

	context, err := zmq.NewContext()

	if err != nil {

		log.Println("Error initializing provisioning publisher context")

		return nil, err

	}

	socket, err := context.NewSocket(zmq.PUB)

	if err != nil {

		log.Println("Error initializing provisioning publisher socket")

		return nil, err

	}

	err = socket.Bind("tcp://*:7005")

	if err != nil {

		log.Println("Error binding provisioning publisher socket")

		return nil, err

	}

	return &ProvisioningPublisher{

		context: context,

		socket: socket,
	}, nil

}

func (publisher *ProvisioningPublisher) SendUpdate(objectIds []uint32, topic string) error {

	dataBytes := encode(objectIds)

	_, err := publisher.socket.Send(topic+string(dataBytes), 0) // Currently publishing on "" topic.

	if err != nil {

		log.Println("Error sending update: ", err)

		return err
	}

	return nil

}

func (publisher *ProvisioningPublisher) Close() error {

	err := publisher.socket.Close()

	if err != nil {

		log.Println("Error closing socket: ", err)

		return err

	}

	err = publisher.context.Term()

	if err != nil {

		log.Println("Error closing context: ", err)

		return err

	}

	return nil

}

func encode(objectIds []uint32) []byte {

	bytes := make([]byte, 4*len(objectIds))

	for index, objectId := range objectIds {

		binary.LittleEndian.PutUint32(bytes[index*4:], objectId)

	}

	return bytes

}
