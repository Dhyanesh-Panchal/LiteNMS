package services

import (
	zmq "github.com/pebbe/zmq4"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	. "nms-backend/utils"
)

type ProvisioningPublisher struct {
	context *zmq.Context
	socket  *zmq.Socket
}

func InitProvisioningPublisher() (*ProvisioningPublisher, error) {

	context, err := zmq.NewContext()

	if err != nil {

		Logger.Error("Failed to initialize provisioning publisher client", zap.Error(err))

		return nil, err

	}

	socket, err := context.NewSocket(zmq.PUB)

	if err != nil {

		Logger.Error("Failed to initialize provisioning publisher socket", zap.Error(err))

		return nil, err

	}

	err = socket.Bind("tcp://*:" + ProvisionPublisherPort)

	if err != nil {

		Logger.Error("Error binding provisioning publisher socket", zap.Error(err))

		return nil, err

	}

	return &ProvisioningPublisher{

		context: context,

		socket: socket,
	}, nil

}

func (publisher *ProvisioningPublisher) SendUpdate(objectIds []string, topic string) error {

	dataBytes, err := msgpack.Marshal(map[string][]string{

		"updateProvisionIps": objectIds,
	})

	if err != nil {

		Logger.Error("Error marshalling update provision status", zap.Error(err))

		return err

	}

	_, err = publisher.socket.Send(topic+string(dataBytes), 0) // Currently publishing on "" topic.

	if err != nil {

		Logger.Error("Error sending update", zap.Error(err))

		return err
	}

	return nil

}

func (publisher *ProvisioningPublisher) Close() error {

	err := publisher.socket.Close()

	if err != nil {

		Logger.Error("Error closing provisioning publisher socket", zap.Error(err))

		return err

	}

	err = publisher.context.Term()

	if err != nil {

		Logger.Error("Error terminating provisioning publisher context", zap.Error(err))

		return err

	}

	return nil

}
