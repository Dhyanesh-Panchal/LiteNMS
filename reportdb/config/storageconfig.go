package config

import "os"

var (
	PartitionCount      uint32 = 5
	BlockSize           uint32 = 120
	FileSizeGrowthDelta int64  = 10 * int64(os.Getpagesize()) // Percentage
	InitialFileSize     int64  = 5 * int64(os.Getpagesize())
)
