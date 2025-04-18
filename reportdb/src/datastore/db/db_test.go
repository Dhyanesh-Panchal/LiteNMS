package db

//type PollingData struct {
//	PollingData []PolledDataPoint `json:"polling_data"`
//}
//
//func TestReportDB_Write(t *testing.T) {
//	err := LoadConfig()
//
//	if err != nil {
//
//		t.Errorf("Error loading config: %v", err)
//
//	}
//
//	dataWriteChannel := make(chan []PolledDataPoint, DataWriteChannelSize)
//
//	db, err := InitDB(dataWriteChannel)
//
//	if err != nil {
//		t.Error(err)
//	}
//
//	// Read and unmarshal JSON data
//	data, err := os.ReadFile(CurrentWorkingDirectory + "/test-data/polling_data_2.json")
//
//	if err != nil {
//
//		log.Fatalf("Error reading JSON file: %v", err)
//
//	}
//
//	var pollingData PollingData
//
//	if err := json.Unmarshal(data, &pollingData); err != nil {
//
//		log.Fatalf("Error unmarshalling JSON: %v", err)
//
//	}
//
//	fmt.Println(len(pollingData.PollingData))
//
//	//fmt.Println(reflect.TypeOf(pollingData.PollingData[0].Value))
//
//	for range 1 {
//		// Send data points to channel
//		for i := 0; i < len(pollingData.PollingData)/20; i++ {
//
//			db.dataWriteChannel <- pollingData.PollingData[i*20 : (i+1)*20]
//
//			time.Sleep(400 * time.Millisecond)
//
//		}
//	}
//
//	time.Sleep(1 * time.Second)
//
//}
//
//func TestQueryHistogram(t *testing.T) {
//	err := LoadConfig()
//
//	if err != nil {
//
//		t.Errorf("Error loading config: %v", err)
//
//	}
//
//	dataWriteChannel := make(chan []PolledDataPoint, DataWriteChannelSize)
//
//	db, err := InitDB(dataWriteChannel)
//
//	if err != nil {
//
//		t.Errorf("Error initializing DB: %v", err)
//
//	}
//
//	from := uint32(1744191393)
//
//	to := uint32(1744362948)
//
//	data, err := db.QueryHistogram(from, to, 2, []uint32{169093227})
//
//	if err != nil {
//
//		t.Fatal(err)
//
//	}
//
//	log.Println(data)
//
//}
