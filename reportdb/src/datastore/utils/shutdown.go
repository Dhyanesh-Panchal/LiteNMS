package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func InitShutdownHandler(signalCount int) <-chan bool {

	GlobalShutdown := make(chan bool, signalCount)

	osSignal := make(chan os.Signal, 1)

	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)

	// start Listening for signal

	go func(signalCount int) {
		<-osSignal

		// signal received, broadcast shutdown
		for range signalCount {

			GlobalShutdown <- true

		}

		Logger.Info("global shutdown signals sent.")

	}(signalCount)

	return GlobalShutdown

}
