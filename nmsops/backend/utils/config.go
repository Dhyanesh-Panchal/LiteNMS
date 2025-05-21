package utils

import (
	"encoding/json"
	"go.uber.org/zap"
	"log"
	"os"
)

var (
	ServerPort              string
	ConfigDBUser            string
	ConfigDBPassword        string
	ConfigDBName            string
	ConfigDBHost            string
	ConfigDBPort            string
	ReportDBHost            string
	ReportDBQueryPort       string
	ReportDBQueryResultPort string
	ProvisionPublisherPort  string
	PollReceiverPort        string
	PollSenderPort          string
	PollDataChannelSize     int
	QuerySendChannelSize    int
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
	ServerPort = generalConfig["ServerPort"].(string)

	ConfigDBUser = generalConfig["ConfigDBUser"].(string)

	ConfigDBPassword = generalConfig["ConfigDBPassword"].(string)

	ConfigDBName = generalConfig["ConfigDBName"].(string)

	ConfigDBHost = generalConfig["ConfigDBHost"].(string)

	ConfigDBPort = generalConfig["ConfigDBPort"].(string)

	ReportDBHost = generalConfig["ReportDBHost"].(string)

	ReportDBQueryPort = generalConfig["ReportDBQueryPort"].(string)

	ReportDBQueryResultPort = generalConfig["ReportDBQueryResultPort"].(string)

	ProvisionPublisherPort = generalConfig["ProvisionPublisherPort"].(string)

	PollReceiverPort = generalConfig["PollReceiverPort"].(string)

	PollSenderPort = generalConfig["PollSenderPort"].(string)

	PollDataChannelSize = int(generalConfig["PollDataChannelSize"].(float64))

	QuerySendChannelSize = int(generalConfig["QuerySendChannelSize"].(float64))

	MaxLogFileSizeInMB = int(generalConfig["MaxLogFileSizeInMB"].(float64))

	LogFileRetentionInDays = int(generalConfig["LogFileRetentionInDays"].(float64))

	IsProductionEnvironment = generalConfig["IsProductionEnvironment"].(bool)

	return nil

}

func GetConfigDBConnectionString() string {
	return "postgres://" + ConfigDBUser + ":" + ConfigDBPassword + "@" + ConfigDBHost + ":" + ConfigDBPort + "/" + ConfigDBName + "?sslmode=disable"
}
