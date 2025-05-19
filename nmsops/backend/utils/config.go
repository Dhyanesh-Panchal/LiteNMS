package utils

import (
	"encoding/json"
	"go.uber.org/zap"
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
)

func LoadConfig() error {

	currentWorkingDirectory, _ := os.Getwd()

	configFilesDir := currentWorkingDirectory + "/config"

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

	return nil

}

func GetConfigDBConnectionString() string {
	return "postgres://" + ConfigDBUser + ":" + ConfigDBPassword + "@" + ConfigDBHost + ":" + ConfigDBPort + "/" + ConfigDBName + "?sslmode=disable"
}
