package config

const (
	PartitionCount      uint32 = 5
	BlockSize           uint32 = 120
	FileSizeGrowthDelta uint32 = 10 * 4096 // Percentage
	InitialFileSize     uint32 = 4096
)
