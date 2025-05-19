package services

import (
	"errors"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	. "nms-backend/db"
	. "nms-backend/utils"
)

type PollDataListener struct {
	context  *zmq.Context
	shutdown chan struct{}
}

func InitPollDataListener(reportDB *ReportDBClient) *PollDataListener {

	context, err := zmq.NewContext()

	if err != nil {

		Logger.Error("Failed to initialize provisioning publisher client", zap.Error(err))

		return nil

	}

	pollDataListener := PollDataListener{

		context: context,

		shutdown: make(chan struct{}, 1),
	}

	go pollDataListener.listener(context, reportDB)

	return &pollDataListener

}

func (pollDataListener *PollDataListener) listener(context *zmq.Context, reportDB *ReportDBClient) {

	// Receiver socket listening for poll data from polling-engine

	receiver, err := context.NewSocket(zmq.PULL)

	if err != nil {

		Logger.Error("Failed to initialize poll listener socket", zap.Error(err))

		return

	}

	if err = receiver.Bind("tcp://*:" + PollReceiverPort); err != nil {

		Logger.Error("Failed to bind", zap.Error(err))

		return

	}

	for {

		select {

		case <-pollDataListener.shutdown:

			if err = receiver.Close(); err != nil {

				Logger.Error("Failed to close poll data receiver socket", zap.Error(err))

			}

			// acknowledge
			pollDataListener.shutdown <- struct{}{}

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

			reportDB.SendPollData(polledData)
		}

	}

}

func (pollDataListener *PollDataListener) Close() error {

	pollDataListener.shutdown <- struct{}{}

	err := pollDataListener.context.Term()

	if err != nil {

		Logger.Error("Failed to terminate poll data router context", zap.Error(err))

		return err
	}

	// Wait for sockets to close

	<-pollDataListener.shutdown

	return nil

}
