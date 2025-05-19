package utils

import (
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func InitProfiling() {

	if !IsProductionEnvironment {

		// Log processID for debug purposes.
		Logger.Debug("Process ID: ", zap.Int("id", os.Getpid()))

		if err := http.ListenAndServe("localhost:"+ProfilingPort, nil); err != nil {

			Logger.Error("error starting profiling server", zap.Error(err))

		}

	}

}
