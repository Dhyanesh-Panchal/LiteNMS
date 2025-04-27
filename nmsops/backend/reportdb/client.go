package reportdb

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"log"
	"sync/atomic"
	"time"
)

var (
	ErrCreatingSocket = errors.New("error creating socket")
	ErrQueryTimedOut  = errors.New("query timed out")
)

type Query struct {
	QueryId               uint64   `json:"query_id"`
	From                  uint32   `json:"from"`
	To                    uint32   `json:"to"`
	ObjectIds             []uint32 `json:"object_ids"`
	CounterId             uint16   `json:"counter_id"`
	VerticalAggregation   string   `json:"vertical_aggregation"`
	HorizontalAggregation string   `json:"horizontal_aggregation"`
	Interval              uint32   `json:"interval"`
}

type DataPoint struct {
	Timestamp uint32      `json:"timestamp"`
	Value     interface{} `json:"value"`
}

type Result struct {
	QueryId uint64 `json:"query_id"`

	Data map[uint32][]DataPoint `json:"data"`
}

type ReportDBClient struct {
	context *zmq.Context

	receiverWaitChannels map[uint64]chan []byte

	queryChannel chan []byte

	shutdownChannel chan struct{}

	queryId uint64
}

func InitClient() (*ReportDBClient, error) {

	context, err := zmq.NewContext()

	if err != nil {

		return nil, err

	}

	receiverWaitChannels := make(map[uint64]chan []byte)

	querySendChannel := make(chan []byte)

	shutdownChannel := make(chan struct{})

	go querySendRoutine(context, querySendChannel)

	go resultReceiveRoutine(context, receiverWaitChannels, shutdownChannel)

	return &ReportDBClient{context, receiverWaitChannels, querySendChannel, shutdownChannel, 0}, nil

}

func querySendRoutine(context *zmq.Context, queryChannel chan []byte) {

	socket, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		log.Println(ErrCreatingSocket)

		return

	}

	defer socket.Close()

	err = socket.Connect("tcp://localhost:7001")

	for query := range queryChannel {

		_, err := socket.SendBytes(query, 0)

		if err != nil {

			log.Println("Error sending query", err)

		}

	}

	log.Println("Query send routine closed")
}

func resultReceiveRoutine(context *zmq.Context, receiverWaitChannels map[uint64]chan []byte, shutdown chan struct{}) {

	socket, err := context.NewSocket(zmq.PULL)

	if err != nil {

		log.Println("Error creating socket", err)

		return
	}

	err = socket.Connect("tcp://localhost:7002")

	if err != nil {

		log.Println(ErrCreatingSocket)

	}

	for {
		select {

		case <-shutdown:

			socket.Close()

			// Acknowledge
			shutdown <- struct{}{}

			return

		default:

			resultBytes, err := socket.RecvBytes(0)

			if err != nil {

				log.Println("Error receiving query", err)

				continue
			}

			queryId := binary.LittleEndian.Uint64(resultBytes[:8])

			fmt.Println("Result received for queryId ", queryId)

			fmt.Printf("receiver wait channels map from receiver routine: %p\n", &receiverWaitChannels)

			if channel, ok := receiverWaitChannels[queryId]; ok {
				fmt.Println("Result Published on channel for queryId ", queryId)

				channel <- resultBytes[8:]

				close(channel)

				delete(receiverWaitChannels, queryId)

			} else {

				fmt.Println("No channel on receiverChannels for queryId ", queryId, "receiverMap", receiverWaitChannels)

			}

		}
	}

}

func (db *ReportDBClient) Query(from, to, interval uint32, objectIds []uint32, counterId uint16, verticalAggregation, horizontalAggregation string) (map[uint32][]DataPoint, error) {

	queryId := atomic.SwapUint64(&db.queryId, db.queryId+1)

	queryBytes, err := json.Marshal(Query{
		QueryId:               queryId,
		From:                  from,
		To:                    to,
		ObjectIds:             objectIds,
		CounterId:             counterId,
		VerticalAggregation:   verticalAggregation,
		HorizontalAggregation: horizontalAggregation,
		Interval:              interval,
	})

	if err != nil {

		log.Println("Error marshalling query", err)

		return nil, err

	}

	receiveChannel := make(chan []byte)

	db.receiverWaitChannels[queryId] = receiveChannel

	fmt.Printf("On the Query Side %p\n", &db.receiverWaitChannels)

	// Send query
	db.queryChannel <- queryBytes

	fmt.Println("Query sent for queryID ", queryId)

	select {

	case <-time.After(10 * time.Second):

		log.Println("Timeout while receiving query")

		return nil, ErrQueryTimedOut

	case resultBytes := <-receiveChannel:

		var result Result

		if err = json.Unmarshal(resultBytes, &result); err != nil {

			log.Println("Error unmarshalling result", err)

			return nil, err
		}

		return result.Data, nil

	}

}

func (db *ReportDBClient) Shutdown() {

	close(db.queryChannel)

	db.shutdownChannel <- struct{}{}

	db.context.Term()

	<-db.shutdownChannel

	for _, channel := range db.receiverWaitChannels {

		close(channel)

	}

}
