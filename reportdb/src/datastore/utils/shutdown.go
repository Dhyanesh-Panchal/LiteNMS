package utils

import (
	"fmt"
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

			fmt.Println("Shutdown sent")

			GlobalShutdown <- true

		}

		log.Println(signalCount, "Shutdown signals sent.")

	}(signalCount)

	return GlobalShutdown

}
