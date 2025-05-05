package services

import (
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	. "nms-backend/utils"
)

func InitPollRouter() *zmq.Context {

	context, err := zmq.NewContext()

	if err != nil {

		Logger.Error("Failed to initialize provisioning publisher client", zap.Error(err))

		return nil

	}

	// TODO: Handle shutdown

	go routerSchedule(context)

	return context

}

func routerSchedule(context *zmq.Context) {

	// Receiver socket listening for poll data from polling-engine

	receiver, err := context.NewSocket(zmq.PULL)

	if err != nil {

		Logger.Error("Failed to initialize provisioning publisher socket", zap.Error(err))

		return

	}

	defer receiver.Close()

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

	defer sender.Close()

	if err = sender.Connect("tcp://" + ReportDBHost + ":" + PollSenderPort); err != nil {

		Logger.Error("Failed to bind", zap.Error(err))

		return

	}

	for {

		polledData, err := receiver.RecvBytes(0)

		if err != nil {

			Logger.Error("Failed to receive polled Data", zap.Error(err))

			continue

		}

		_, err = sender.SendBytes(polledData, 0)

		if err != nil {

			Logger.Error("Failed to send polled Data", zap.Error(err))

			continue

		}

	}

}
