package reportdb

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reportdb/containers"
	"testing"
	"time"
)

type PollingData struct {
	PollingData []containers.PolledDataPoint `json:"polling_data"`
}

func TestReportDB_Write(t *testing.T) {
	db := InitDB()

	// Read and unmarshal JSON data
	data, err := os.ReadFile("polling_data_2.json")

	if err != nil {

		log.Fatalf("Error reading JSON file: %v", err)

	}

	var pollingData PollingData

	if err := json.Unmarshal(data, &pollingData); err != nil {

		log.Fatalf("Error unmarshaling JSON: %v", err)

	}

	fmt.Println(len(pollingData.PollingData))

	// Send data points to channel
	for i := 0; i < len(pollingData.PollingData)/20; i++ {

		db.Write(pollingData.PollingData[i*20 : (i+1)*20])

		time.Sleep(400 * time.Millisecond)

	}

	time.Sleep(1 * time.Second)

}
