package server

import (
	"errors"
	zmq "github.com/pebbe/zmq4"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	. "poller/containers"
	. "poller/utils"
	"sync"
)

func InitProvisionListener(deviceList *DeviceList, globalShutdown <-chan struct{}, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	context, err := zmq.NewContext()

	if err != nil {

		Logger.Error("Error creating zmq context", zap.Error(err))

		return

	}

	provisionListenerShutdown := make(chan struct{}, 1)

	go provisionListener(deviceList, provisionListenerShutdown)

	// Listen for global shutdown
	<-globalShutdown

	// Send shutdown to socket
	provisionListenerShutdown <- struct{}{}

	err = context.Term()

	if err != nil {

		Logger.Error("error terminating query listener context", zap.Error(err))

		return

	}

	// Wait for socket to close.
	<-provisionListenerShutdown

}

func provisionListener(deviceList *DeviceList, provisionListenerShutdown chan struct{}) {

	socket, err := zmq.NewSocket(zmq.SUB)

	if err != nil {

		Logger.Fatal("Error creating zmq socket for provision listener", zap.Error(err))

	}

	err = socket.Connect("tcp://" + BackendHost + ":" + ProvisionListenerPort)

	if err != nil {

		Logger.Fatal("Error binding the socket", zap.Error(err))

	}

	if err = socket.SetSubscribe(""); err != nil {

		Logger.Fatal("Error setting subscribe", zap.Error(err))

		return
	}

	for {

		select {

		case <-provisionListenerShutdown:

			err := socket.Close()

			if err != nil {

				Logger.Error("error closing query listener socket ", zap.Error(err))

			}

			// Acknowledge
			provisionListenerShutdown <- struct{}{}

			return

		default:
			responseBytes, err := socket.RecvBytes(0)

			if err != nil {

				if errors.Is(zmq.AsErrno(err), zmq.ETERM) {

					Logger.Info("Provision Listener's ZMQ-Context terminated, closing the socket")

				} else {

					Logger.Error("error receiving provision update ", zap.Error(err))

				}

				continue

			}

			var provisionUpdateIps map[string][]string

			err = msgpack.Unmarshal(responseBytes, &provisionUpdateIps)

			if err != nil {

				Logger.Error("error decoding the provision Update ", zap.Error(err))

			}

			deviceList.UpdateProvisionedDeviceList(provisionUpdateIps["updateProvisionIps"])

			Logger.Info("Updated the device provisioning list", zap.Any("provisionUpdate", provisionUpdateIps))

		}

	}

}
