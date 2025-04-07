package config

import (
	"time"
)

const (
	ProjectRootPath     = "/home/dhyanesh/Desktop/LiteNMS/reportdb"
	WriterCount         = 5
	WriterFlushDuration = time.Second * 1
)

var CounterConfig = map[uint16]map[string]interface{}{
	1: {
		"dataType": "int64",
		"dataSize": uint32(8),
	},
	2: {
		"dataType": "float64",
		"dataSize": uint32(8),
	},
	3: {
		"dataType": "string",
	},
}
