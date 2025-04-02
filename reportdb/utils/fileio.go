package utils

import (
	"encoding/json"
	"os"
)

func LoadJsonFile(filePath string, targetContainer *interface{}) error {

	data, err := os.ReadFile(filePath)

	if err != nil {

		return err

	}

	err = json.Unmarshal(data, targetContainer)

	if err != nil {

		return err

	}
	return nil

}
