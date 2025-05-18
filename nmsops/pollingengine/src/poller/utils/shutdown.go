package utils

import (
	"os"
	"os/signal"
	"syscall"
)

func InitShutdownHandler(signalCount int) <-chan struct{} {

	GlobalShutdown := make(chan struct{}, signalCount)

	osSignal := make(chan os.Signal, 1)

	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)

	// start Listening for signal

	go func(signalCount int) {
		<-osSignal

		// signal received, broadcast shutdown
		for range signalCount {

			GlobalShutdown <- struct{}{}

		}

		Logger.Info("global shutdown signals sent.")

	}(signalCount)

	return GlobalShutdown

}
