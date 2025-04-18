package utils

import (
	"go.uber.org/zap"
	"net/http"
	_ "net/http/pprof"
)

func InitProfiling() {
	err := http.ListenAndServe("localhost:"+ProfilingPort, nil)

	if err != nil {

		Logger.Error("error starting profiling server", zap.Error(err))

	}
}
