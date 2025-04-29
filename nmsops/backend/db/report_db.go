package db

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	. "nms-backend/utils"
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

func InitReportDBClient() (*ReportDBClient, error) {

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

		Logger.Error("Error creating zmq socket", zap.Error(err))

		return

	}

	defer socket.Close()

	err = socket.Connect("tcp://localhost:7001")

	for query := range queryChannel {

		_, err := socket.SendBytes(query, 0)

		if err != nil {

			Logger.Error("Error sending query", zap.Error(err))

		}

	}

	Logger.Info("Query sender routine closed")
}

func resultReceiveRoutine(context *zmq.Context, receiverWaitChannels map[uint64]chan []byte, shutdown chan struct{}) {

	socket, err := context.NewSocket(zmq.PULL)

	if err != nil {

		Logger.Error("Error creating zmq socket", zap.Error(err))

		return
	}

	err = socket.Connect("tcp://localhost:7002")

	if err != nil {

		Logger.Error("Error connecting the socket", zap.Error(err))

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

				if errors.Is(zmq.AsErrno(err), zmq.ETERM) {

					Logger.Info("Result receiver's ZMQ-Context terminated, closing the socket")

				} else {

					Logger.Error("error receiving query ", zap.Error(err))

				}

				continue

			}

			queryId := binary.LittleEndian.Uint64(resultBytes[:8])

			if channel, ok := receiverWaitChannels[queryId]; ok {

				channel <- resultBytes[8:]

				close(channel)

				delete(receiverWaitChannels, queryId)

			}

		}
	}

}

func (db *ReportDBClient) Query(from, to, interval uint32, objectIps []string, counterId uint16, verticalAggregation, horizontalAggregation string) (interface{}, error) {

	queryId := atomic.SwapUint64(&db.queryId, db.queryId+1)

	objectIds := make([]uint32, len(objectIps))

	for index, ip := range objectIps {

		objectIds[index] = ConvertIpToNumeric(ip)

	}

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

		Logger.Error("Error serializing query", zap.Error(err))

		return nil, err

	}

	db.receiverWaitChannels[queryId] = make(chan []byte)

	// Send query
	db.queryChannel <- queryBytes

	select {

	case <-time.After(10 * time.Second):

		Logger.Error("Query timeout", zap.Uint64("queryId", queryId))

		return nil, ErrQueryTimedOut

	case resultBytes := <-db.receiverWaitChannels[queryId]:

		var result Result

		if err = json.Unmarshal(resultBytes, &result); err != nil {

			Logger.Error("Error deserializing query result", zap.Error(err))

			return nil, err
		}

		return parseResponse(result.Data), nil

	}

}

func parseResponse(data map[uint32][]DataPoint) interface{} {

	if result, exist := data[0]; exist {

		// Result of Query without groupBy
		// Hence return the single result array.

		return result

	} else {

		// query with groupBy over objectIds
		// Convert objectIds to string and return the map.

		response := make(map[string][]DataPoint)

		for objectId, result := range data {

			response[ConvertNumericToIp(objectId)] = result

		}

		return response
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
