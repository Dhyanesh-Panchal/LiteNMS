package server

import (
	"encoding/binary"
	"errors"
	zmq "github.com/pebbe/zmq4"
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

	socket.SetSubscribe("")

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

			provisionUpdateIps := make([]uint32, len(responseBytes)/4)

			err = decode(responseBytes, provisionUpdateIps)

			if err != nil {

				Logger.Error("error decoding the provision Update ", zap.Error(err))

			}

			deviceList.UpdateProvisionedDeviceList(provisionUpdateIps)

			Logger.Info("Updated the device provisioning list", zap.Any("provisionUpdate", provisionUpdateIps))

		}

	}

}

func decode(dataBytes []byte, data []uint32) (err error) {

	err = nil

	defer func() {
		if r := recover(); r != nil {

			Logger.Error("Panic in decoder", zap.Any("recover", r))

			err = r.(error)

		}
	}()

	for i := 0; i < len(dataBytes)/4; i++ {

		data[i] = binary.LittleEndian.Uint32(dataBytes[i*4 : (i+1)*4])

	}

	return err

}
