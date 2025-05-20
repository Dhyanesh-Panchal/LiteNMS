package utils

import (
	"encoding/json"
	"errors"
	"github.com/bytedance/gopkg/util/gctuner"
	"go.uber.org/zap"
	"log"
	"os"
	"syscall"
)

const DataType = "dataType"

var CounterConfig map[uint16]map[string]string

var (
	Writers                   int
	DataWriteChannelSize      int
	Readers                   int
	ReaderRequestChannelSize  int
	ReaderResponseChannelSize int
	QueryParsers              int
	QueryChannelSize          int
	QueryTimeoutTime          int
	Partitions                uint32
	BlockSize                 uint32
	FileSizeGrowthDelta       int64
	InitialFileSize           int64
	StorageCleanupInterval    int
	MaxCacheKeys              int64
	MaxCacheSizeInMB          int64
	PollListenerBindPort      string
	QueryListenerBindPort     string
	QueryResultBindPort       string
	ProfilingPort             string
	StorageDirectory          string
	IsProductionEnvironment   bool
	MaxLogFileSizeInMB        int
	LogFileRetentionInDays    int
)

func LoadConfig() (err error) {

	defer func() {

		if r := recover(); r != nil {

			log.Println("Invalid config: ", r)

			err = r.(error)

		}

	}()

	currentWorkingDirectory, _ := os.Getwd()

	StorageDirectory = currentWorkingDirectory + "/data"

	configFilesDir := currentWorkingDirectory + "/config"

	countersConfigBytes, err := os.ReadFile(configFilesDir + "/counters.json")

	if err != nil {

		log.Println("Unable to read counter file: ", zap.Error(err))

		return err

	}

	if err = json.Unmarshal(countersConfigBytes, &CounterConfig); err != nil {

		log.Println("Unable to unmarshal counter config data: ", zap.Error(err))

		return err

	}

	generalConfigBytes, err := os.ReadFile(configFilesDir + "/general.json")

	if err != nil {

		log.Println("Unable to read general config file: ", zap.Error(err))

		return err

	}

	var generalConfig map[string]interface{}

	if err = json.Unmarshal(generalConfigBytes, &generalConfig); err != nil {

		log.Println("Unable to unmarshal general config data: ", zap.Error(err))

		return err

	}

	// Set General Config Variables
	Writers = int(generalConfig["Writers"].(float64))

	DataWriteChannelSize = int(generalConfig["DataWriteChannelSize"].(float64))

	Readers = int(generalConfig["Readers"].(float64))

	ReaderRequestChannelSize = int(generalConfig["ReaderRequestChannelSize"].(float64))

	ReaderResponseChannelSize = int(generalConfig["ReaderResponseChannelSize"].(float64))

	QueryParsers = int(generalConfig["QueryParsers"].(float64))

	QueryChannelSize = int(generalConfig["QueryChannelSize"].(float64))

	QueryTimeoutTime = int(generalConfig["QueryTimeoutTime"].(float64))

	Partitions = uint32(generalConfig["Partitions"].(float64))

	BlockSize = uint32(generalConfig["BlockSize"].(float64))

	pageSize := int64(os.Getpagesize())

	InitialFileSize = int64(generalConfig["InitialFileSize"].(float64)) * pageSize

	FileSizeGrowthDelta = int64(generalConfig["FileSizeGrowthDelta"].(float64)) * pageSize

	StorageCleanupInterval = int(generalConfig["StorageCleanupInterval"].(float64))

	MaxCacheKeys = int64(generalConfig["MaxCacheKeys"].(float64))

	MaxCacheSizeInMB = int64(generalConfig["MaxCacheSizeInMB"].(float64))

	PollListenerBindPort = generalConfig["PollListenerBindPort"].(string)

	QueryListenerBindPort = generalConfig["QueryListenerBindPort"].(string)

	QueryResultBindPort = generalConfig["QueryResultBindPort"].(string)

	ProfilingPort = generalConfig["ProfilingPort"].(string)

	IsProductionEnvironment = generalConfig["IsProductionEnvironment"].(bool)

	MaxLogFileSizeInMB = int(generalConfig["MaxLogFileSizeInMB"].(float64))

	LogFileRetentionInDays = int(generalConfig["LogFileRetentionInDays"].(float64))

	// Validate the config variables

	if err = validateConfigParameters(); err != nil {

		return err

	}

	//Get system memory and set GC tuning
	memoryThreshold := (sysTotalMemory() * uint64(generalConfig["MemoryFraction"].(float64))) / 100

	gctuner.Tuning(memoryThreshold)

	return nil

}

func validateConfigParameters() error {

	if Writers <= 0 {
		return errors.New("Writers should be greater than 0")
	}

	if DataWriteChannelSize <= 0 {
		return errors.New("DataWriteChannelSize should be greater than 0")
	}

	if Readers <= 0 {
		return errors.New("Readers should be greater than 0")
	}

	if ReaderRequestChannelSize <= 0 {
		return errors.New("ReaderRequestChannelSize should be greater than 0")
	}

	if ReaderResponseChannelSize <= 0 {
		return errors.New("ReaderResponseChannelSize should be greater than 0")
	}

	if QueryParsers <= 0 {
		return errors.New("QueryParsers should be greater than 0")
	}

	if QueryChannelSize <= 0 {
		return errors.New("QueryChannelSize should be greater than 0")
	}

	if QueryTimeoutTime <= 0 {
		return errors.New("QueryTimeoutTime should be greater than 0")
	}

	if Partitions <= 0 {
		return errors.New("Partitions should be greater than 0")
	}

	if BlockSize <= 0 {
		return errors.New("BlockSize should be greater than 0")
	}

	if FileSizeGrowthDelta <= 0 {
		return errors.New("FileSizeGrowthDelta should be greater than 0")
	}

	if InitialFileSize <= 0 {
		return errors.New("InitialFileSize should be greater than 0")
	}

	if StorageCleanupInterval <= 0 {
		return errors.New("StorageCleanupInterval should be greater than 0")
	}

	if MaxCacheKeys <= 0 {
		return errors.New("MaxCacheKeys should be greater than 0")
	}

	if MaxCacheSizeInMB <= 0 {
		return errors.New("MaxCacheSizeInMB should be greater than 0")
	}

	if MaxLogFileSizeInMB <= 0 {
		return errors.New("MaxLogFileSizeInMB should be greater than 0")
	}

	if LogFileRetentionInDays <= 0 {
		return errors.New("LogFileRetentionInDays should be greater than 0")
	}

	if PollListenerBindPort == "" {
		return errors.New("PollListenerBindPort should not be empty")
	}

	if QueryListenerBindPort == "" {
		return errors.New("QueryListenerBindPort should not be empty")
	}

	if QueryResultBindPort == "" {
		return errors.New("QueryResultBindPort should not be empty")
	}

	if ProfilingPort == "" {
		return errors.New("ProfilingPort should not be empty")
	}

	return nil

}

func sysTotalMemory() uint64 {

	in := &syscall.Sysinfo_t{}

	err := syscall.Sysinfo(in)

	if err != nil {

		return 0

	}

	// If this is a 32-bit system, then these fields are
	// uint32 instead of uint64.
	// So we always convert to uint64 to match signature.
	return uint64(in.Totalram) * uint64(in.Unit)
}
