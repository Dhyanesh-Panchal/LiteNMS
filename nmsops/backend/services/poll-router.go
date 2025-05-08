package services

import (
	"errors"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	. "nms-backend/utils"
)

type PollDataRouter struct {
	context  *zmq.Context
	shutdown chan struct{}
}

func InitPollDataRouter() *PollDataRouter {

	context, err := zmq.NewContext()

	if err != nil {

		Logger.Error("Failed to initialize provisioning publisher client", zap.Error(err))

		return nil

	}

	// TODO: Handle shutdown

	pollDataRouter := PollDataRouter{

		context: context,

		shutdown: make(chan struct{}, 1),
	}

	go pollDataRouter.routerSchedule(context)

	return &pollDataRouter

}

func (pollDataRouter *PollDataRouter) routerSchedule(context *zmq.Context) {

	// Receiver socket listening for poll data from polling-engine

	receiver, err := context.NewSocket(zmq.PULL)

	if err != nil {

		Logger.Error("Failed to initialize provisioning publisher socket", zap.Error(err))

		return

	}

	if err = receiver.Bind("tcp://*:" + PollReceiverPort); err != nil {

		Logger.Error("Failed to bind", zap.Error(err))

		return

	}

	// Sender socket routing the polled data to reportDB

	sender, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		Logger.Error("Failed to initialize provisioning publisher socket", zap.Error(err))

		return

	}

	sender.SetLinger(0)

	if err = sender.Connect("tcp://" + ReportDBHost + ":" + PollSenderPort); err != nil {

		Logger.Error("Failed to bind", zap.Error(err))

		return

	}

	for {

		select {

		case <-pollDataRouter.shutdown:

			sender.Close()

			receiver.Close()

			Logger.Info("Poll data router sockets closed.")

			// acknowledge
			pollDataRouter.shutdown <- struct{}{}

			return

		default:

			polledData, err := receiver.RecvBytes(0)

			if err != nil {

				if errors.Is(zmq.AsErrno(err), zmq.ETERM) {

					Logger.Info("Poll data router's ZMQ-Context terminated, closing the sockets")

				} else {

					Logger.Error("Failed to receive polled Data", zap.Error(err))

				}

				continue

			}

			_, err = sender.SendBytes(polledData, 0)

			if err != nil {

				Logger.Error("Failed to send polled Data", zap.Error(err))

				continue

			}
		}

	}

}

func (pollDataRouter *PollDataRouter) Close() error {

	pollDataRouter.shutdown <- struct{}{}

	err := pollDataRouter.context.Term()

	if err != nil {

		Logger.Error("Failed to terminate poll data router context", zap.Error(err))

		return err
	}

	// Wait for sockets to close

	<-pollDataRouter.shutdown

	return nil

}
