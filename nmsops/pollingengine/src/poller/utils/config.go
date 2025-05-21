package utils

import (
	"encoding/json"
	"go.uber.org/zap"
	"log"
	"os"
)

var CounterConfig = map[uint16]map[string]interface{}{}

var (
	PollSenderPort          string
	BackendHost             string
	ProvisionListenerPort   string
	PollWorkers             int
	PollChannelSize         int
	PollDataBatchSize       int
	ConfigDBUser            string
	ConfigDBPassword        string
	ConfigDBHost            string
	ConfigDBPort            string
	ConfigDBName            string
	CommandDelimiter        string
	MaxLogFileSizeInMB      int
	LogFileRetentionInDays  int
	IsProductionEnvironment bool
)

func LoadConfig() (err error) {

	defer func() {

		if r := recover(); r != nil {

			log.Println("Invalid Config: ", r)

			err = r.(error)

		}

	}()

	currentWorkingDirectory, _ := os.Getwd()

	configFilesDir := currentWorkingDirectory + "/config"

	countersConfigData, err := os.ReadFile(configFilesDir + "/counters.json")

	if err != nil {

		log.Println("Unable to read counter file: ", zap.Error(err))

		return err

	}

	if err = json.Unmarshal(countersConfigData, &CounterConfig); err != nil {

		log.Println("Unable to unmarshal counter config data: ", zap.Error(err))

		return err

	}

	generalConfigData, err := os.ReadFile(configFilesDir + "/general.json")

	if err != nil {

		log.Println("Unable to read general config file: ", zap.Error(err))

		return err

	}

	var generalConfig map[string]interface{}

	if err = json.Unmarshal(generalConfigData, &generalConfig); err != nil {

		log.Println("Unable to unmarshal general config data: ", zap.Error(err))

		return err

	}

	// Set General Config Variables
	PollSenderPort = generalConfig["PollSenderPort"].(string)

	BackendHost = generalConfig["BackendHost"].(string)

	ProvisionListenerPort = generalConfig["ProvisionListenerPort"].(string)

	PollWorkers = int(generalConfig["PollWorkers"].(float64))

	PollChannelSize = int(generalConfig["PollChannelSize"].(float64))

	PollDataBatchSize = int(generalConfig["PollDataBatchSize"].(float64))

	ConfigDBUser = generalConfig["ConfigDBUser"].(string)

	ConfigDBPassword = generalConfig["ConfigDBPassword"].(string)

	ConfigDBHost = generalConfig["ConfigDBHost"].(string)

	ConfigDBPort = generalConfig["ConfigDBPort"].(string)

	ConfigDBName = generalConfig["ConfigDBName"].(string)

	CommandDelimiter = generalConfig["CommandDelimiter"].(string)

	MaxLogFileSizeInMB = int(generalConfig["MaxLogFileSizeInMB"].(float64))

	LogFileRetentionInDays = int(generalConfig["LogFileRetentionInDays"].(float64))

	IsProductionEnvironment = generalConfig["IsProductionEnvironment"].(bool)

	return nil

}

func GetConfigDBConnectionString() string {
	return "postgres://" + ConfigDBUser + ":" + ConfigDBPassword + "@" + ConfigDBHost + ":" + ConfigDBPort + "/" + ConfigDBName + "?sslmode=disable"
}
