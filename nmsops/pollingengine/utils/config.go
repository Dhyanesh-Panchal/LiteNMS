package utils

import (
	"encoding/json"
	"go.uber.org/zap"
	"os"
)

var CounterConfig = map[uint16]map[string]interface{}{}

var (
	PollSenderPort         string
	ProvisionListenerPort  string
	PollWorkers            int
	PollChannelSize        int
	DeviceSSHClientTimeout int
	ConfigDBUser           string
	ConfigDBPassword       string
	ConfigDBHost           string
	ConfigDBPort           string
	ConfigDBName           string
)

func LoadConfig() error {

	currentWorkingDirectory, _ := os.Getwd()

	configFilesDir := currentWorkingDirectory + "/config"

	countersConfigData, err := os.ReadFile(configFilesDir + "/counters.json")

	if err != nil {

		Logger.Info("Unable to read counter file: ", zap.Error(err))

		return err

	}

	if err = json.Unmarshal(countersConfigData, &CounterConfig); err != nil {

		Logger.Info("Unable to unmarshal counter config data: ", zap.Error(err))

		return err

	}

	generalConfigData, err := os.ReadFile(configFilesDir + "/general.json")

	if err != nil {

		Logger.Info("Unable to read general config file: ", zap.Error(err))

		return err

	}

	var generalConfig map[string]interface{}

	if err = json.Unmarshal(generalConfigData, &generalConfig); err != nil {

		Logger.Info("Unable to unmarshal general config data: ", zap.Error(err))

		return err

	}

	// Set General Config Variables
	PollSenderPort = generalConfig["PollSenderPort"].(string)

	ProvisionListenerPort = generalConfig["ProvisionListenerPort"].(string)

	PollWorkers = int(generalConfig["PollWorkers"].(float64))

	PollChannelSize = int(generalConfig["PollChannelSize"].(float64))

	DeviceSSHClientTimeout = int(generalConfig["DeviceSSHClientTimeout"].(float64))

	ConfigDBUser = generalConfig["ConfigDBUser"].(string)

	ConfigDBPassword = generalConfig["ConfigDBPassword"].(string)

	ConfigDBHost = generalConfig["ConfigDBHost"].(string)

	ConfigDBPort = generalConfig["ConfigDBPort"].(string)

	ConfigDBName = generalConfig["ConfigDBName"].(string)

	return nil

}
