package utils

import (
	"encoding/json"
	"log"
	"os"
)

const DataType = "dataType"

var CounterConfig = map[uint16]map[string]interface{}{}

var (
	Writers             int
	BlockSize           uint32
	Partitions          uint32
	InitialFileSize     int64
	FileSizeGrowthDelta int64
)

var CurrentWorkingDirectory string

var StorageDirectory string

func LoadConfig() error {

	CurrentWorkingDirectory, _ = os.Getwd()

	StorageDirectory = CurrentWorkingDirectory + "/data"

	configFilesDir := CurrentWorkingDirectory + "/config"

	countersConfigData, err := os.ReadFile(configFilesDir + "/counters.json")

	if err != nil {

		log.Println("Unable to read counter file: ", err)

		return err

	}

	err = json.Unmarshal(countersConfigData, &CounterConfig)

	if err != nil {

		log.Println("Unable to unmarshal counter config data: ", err)

		return err

	}

	generalConfigData, err := os.ReadFile(configFilesDir + "/general.json")

	if err != nil {

		log.Println("Unable to read general config file: ", err)

		return err

	}

	var generalConfig map[string]interface{}

	err = json.Unmarshal(generalConfigData, &generalConfig)

	if err != nil {

		log.Println("Unable to unmarshal general config data: ", err)

		return err

	}

	// Set General Config Variables
	Writers = int(generalConfig["Writers"].(float64))

	Partitions = uint32(generalConfig["Partitions"].(float64))

	BlockSize = uint32(generalConfig["BlockSize"].(float64))

	pageSize := int64(os.Getpagesize())

	InitialFileSize = int64(generalConfig["InitialFileSize"].(float64)) * pageSize

	FileSizeGrowthDelta = int64(generalConfig["FileSizeGrowthDelta"].(float64)) * pageSize

	return nil

}
