package utils

import (
	"log"
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

		log.Println(signalCount, "Shutdown signals sent.")

	}(signalCount)

	return GlobalShutdown

}
