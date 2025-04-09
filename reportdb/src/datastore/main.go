package main

import (
	. "datastore/containers"
	. "datastore/db"
	. "datastore/utils"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {

	err := LoadConfig()

	if err != nil {

		log.Println("Error loading config:", err)

		return

	}

	database, err := InitDB()

	// Server initialization will come here.

	// Currently Just for testing purpose, reading and writing is performed here.

	// Read and unmarshal JSON data

	data, err := os.ReadFile(CurrentWorkingDirectory + "/test-data/polling_data_2.json")

	if err != nil {

		log.Fatalf("Error reading JSON file: %v", err)

	}

	var pollingData PollingData

	if err := json.Unmarshal(data, &pollingData); err != nil {

		log.Fatal("Error unmarshalling JSON: ", err)

	}

	fmt.Println(len(pollingData.PollingData))

	for i := 0; i < len(pollingData.PollingData)/20; i++ {

		database.Write(pollingData.PollingData[i*20 : (i+1)*20])

		time.Sleep(400 * time.Millisecond)

	}

	time.Sleep(1 * time.Second)

	responce, err := database.QueryHistogram()

}

// PollingData Just for testing
type PollingData struct {
	PollingData []PolledDataPoint `json:"polling_data"`
}
