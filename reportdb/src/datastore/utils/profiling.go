package utils

import (
	"net/http"
	_ "net/http/pprof"
)

func InitProfiling() {
	http.ListenAndServe("localhost:"+ProfilingPort, nil)
}
