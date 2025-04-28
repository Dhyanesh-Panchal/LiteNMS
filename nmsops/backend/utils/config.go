package utils

import (
	"encoding/json"
	"go.uber.org/zap"
	"os"
)

var (
	ConfigDBUser     string
	ConfigDBPassword string
	ConfigDBName     string
	ConfigDBHost     string
	ConfigDBPort     string
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
	ConfigDBUser = generalConfig["ConfigDBUser"].(string)

	ConfigDBPassword = generalConfig["ConfigDBPassword"].(string)

	ConfigDBName = generalConfig["ConfigDBName"].(string)

	ConfigDBHost = generalConfig["ConfigDBHost"].(string)

	ConfigDBPort = generalConfig["ConfigDBPort"].(string)

	return nil

}

func GetConfigDBConnectionString() string {
	return "postgres://" + ConfigDBUser + ":" + ConfigDBPassword + "@" + ConfigDBHost + ":" + ConfigDBPort + "/" + ConfigDBName + "?sslmode=disable"
}
