package reader

import (
	. "datastore/containers"
	. "datastore/storage"
	. "datastore/utils"
	"log"
)

// readSingleDay Note: readFullDate function changes the state of the finalData; hence if run in parallel, proper synchronization is needed.
func readSingleDay(date Date, storageEngine *Storage, counterId uint16, objects []uint32, finalData map[uint32][]DataPoint, from uint32, to uint32) {

	for _, objectId := range objects {

		data, err := storageEngine.Get(objectId)

		if err != nil {

			log.Println("Error getting data for objectId: ", objectId, " Day: ", date)
			continue

		}

		dataPoints, err := DeserializeBatch(data, CounterConfig[counterId][DataType].(string))

		if err != nil {

			log.Println("Error deserializing data for objectId: ", objectId, "Day: ", date, "Error:", err)

			continue

		}

		// Append dataPoints if they lie between from and to

		for _, data := range dataPoints {

			if data.Timestamp >= from && data.Timestamp <= to {

				finalData[objectId] = append(finalData[objectId], data)

			}

		}

	}
}
