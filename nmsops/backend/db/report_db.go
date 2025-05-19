package db

import (
	"encoding/binary"
	"errors"
	zmq "github.com/pebbe/zmq4"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	. "nms-backend/utils"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrQueryTimedOut  = errors.New("query timed out")
	ErrServerShutdown = errors.New("server got shutdown")
)

type Query struct {
	QueryId uint64 `json:"query_id" msgpack:"query_id"`

	From uint32 `json:"from" msgpack:"from"`

	To uint32 `json:"to" msgpack:"to"`

	ObjectIds []uint32 `json:"object_ids" msgpack:"object_ids"`

	CounterId uint16 `json:"counter_id" msgpack:"counter_id"`

	ObjectWiseAggregation string `json:"object_wise_aggregation" msgpack:"object_wise_aggregation"`

	TimestampAggregation string `json:"timestamp_aggregation" msgpack:"timestamp_aggregation"`

	Interval uint32 `json:"interval" msgpack:"interval"`
}

type DataPoint struct {
	Timestamp uint32 `json:"timestamp" msgpack:"timestamp"`

	Value interface{} `json:"value" msgpack:"value"`
}

type Result struct {
	QueryId uint64 `json:"query_id" msgpack:"query_id"`

	Data map[uint32][]DataPoint `json:"data" msgpack:"data"`

	Error string `json:"error" msgpack:"error"`
}

type ReportDBClient struct {
	context *zmq.Context

	receiverWaitChannels map[uint64]chan []byte

	lock sync.RWMutex

	queryChannel chan []byte

	pollDataChannel chan []byte

	shutdownChannel chan struct{}

	queryId uint64
}

func InitReportDBClient() (*ReportDBClient, error) {

	context, err := zmq.NewContext()

	if err != nil {

		return nil, err

	}

	receiverWaitChannels := make(map[uint64]chan []byte)

	querySendChannel := make(chan []byte, QuerySendChannelSize)

	pollDataChannel := make(chan []byte, PollDataChannelSize)

	shutdownChannel := make(chan struct{}, 1)

	client := ReportDBClient{
		context:              context,
		receiverWaitChannels: receiverWaitChannels,
		queryChannel:         querySendChannel,
		pollDataChannel:      pollDataChannel,
		shutdownChannel:      shutdownChannel,
		queryId:              0,
	}

	go querySenderRoutine(context, querySendChannel)

	go resultReceiverRoutine(context, &client, shutdownChannel)

	go pollSenderRoutine(context, pollDataChannel)

	return &client, nil

}

func (db *ReportDBClient) putReceiverChannel(queryId uint64, channel chan []byte) {

	db.lock.Lock()

	defer db.lock.Unlock()

	db.receiverWaitChannels[queryId] = channel

}

func (db *ReportDBClient) getReceiverChannel(queryId uint64) chan []byte {

	db.lock.RLock()

	defer db.lock.RUnlock()

	if channel, ok := db.receiverWaitChannels[queryId]; ok {

		delete(db.receiverWaitChannels, queryId)

		return channel

	} else {

		return nil

	}

}

func (db *ReportDBClient) closeReceivers() {

	db.lock.Lock()

	defer db.lock.Unlock()

	for _, channel := range db.receiverWaitChannels {

		close(channel)

	}

}

func pollSenderRoutine(context *zmq.Context, pollDataChannel chan []byte) {

	socket, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		Logger.Error("Error creating zmq socket", zap.Error(err))

		return

	}

	defer func(socket *zmq.Socket) {

		if err := socket.Close(); err != nil {

			Logger.Error("Error closing poll sender socket", zap.Error(err))

		}

	}(socket)

	if err = socket.SetLinger(0); err != nil {

		Logger.Error("Error setting linger", zap.Error(err))

		return
	}

	if err = socket.Connect("tcp://" + ReportDBHost + ":" + PollSenderPort); err != nil {

		Logger.Error("Failed to bind", zap.Error(err))

		return

	}

	for data := range pollDataChannel {

		_, err := socket.SendBytes(data, 0)

		if err != nil {

			Logger.Error("Error sending polled data", zap.Error(err))

		}

	}

	Logger.Info("Poll sender routine closed")
}

func querySenderRoutine(context *zmq.Context, queryChannel chan []byte) {

	socket, err := context.NewSocket(zmq.PUSH)

	if err != nil {

		Logger.Error("Error creating zmq socket", zap.Error(err))

		return

	}

	if err = socket.SetLinger(0); err != nil {

		Logger.Error("Error setting linger", zap.Error(err))

		return
	}

	defer func(socket *zmq.Socket) {

		if err := socket.Close(); err != nil {

			Logger.Error("Error closing query sender socket", zap.Error(err))

		}

	}(socket)

	err = socket.Connect("tcp://" + ReportDBHost + ":" + ReportDBQueryPort)

	for query := range queryChannel {

		_, err := socket.SendBytes(query, 0)

		if err != nil {

			Logger.Error("Error sending query", zap.Error(err))

		}

	}

	Logger.Info("Query sender routine closed")
}

func resultReceiverRoutine(context *zmq.Context, dbClient *ReportDBClient, shutdown chan struct{}) {

	socket, err := context.NewSocket(zmq.PULL)

	if err != nil {

		Logger.Error("Error creating zmq socket", zap.Error(err))

		return
	}

	err = socket.Connect("tcp://" + ReportDBHost + ":" + ReportDBQueryResultPort)

	if err != nil {

		Logger.Error("Error connecting the socket", zap.Error(err))

	}

	for {
		select {

		case <-shutdown:

			if err = socket.Close(); err != nil {

				Logger.Error("Error closing result receiver socket", zap.Error(err))

				return
			}

			// Acknowledge
			shutdown <- struct{}{}

			Logger.Info("Result receiver routine closed")

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

			receiverChannel := dbClient.getReceiverChannel(queryId)

			receiverChannel <- resultBytes[8:]

			close(receiverChannel)

		}
	}

}

func parseResponse(data map[uint32][]DataPoint) interface{} {

	if result, exist := data[0]; exist {

		// result of query without groupBy
		// hence return the single result array.

		return result

	} else {

		// query with groupBy over objectIds
		// convert objectIds to string IPs and return the map.

		response := make(map[string][]DataPoint)

		for objectId, result := range data {

			response[ConvertNumericToIp(objectId)] = result

		}

		return response
	}

}

func (db *ReportDBClient) Query(from, to, interval uint32, objectIps []string, counterId uint16, objectWiseAggregation, timestampAggregation string) (interface{}, error) {

	queryId := atomic.AddUint64(&db.queryId, 1)

	objectIds := make([]uint32, len(objectIps))

	for index, ip := range objectIps {

		objectIds[index] = ConvertIpToNumeric(ip)

	}

	queryBytes, err := msgpack.Marshal(Query{
		QueryId:               queryId,
		From:                  from,
		To:                    to,
		ObjectIds:             objectIds,
		CounterId:             counterId,
		ObjectWiseAggregation: objectWiseAggregation,
		TimestampAggregation:  timestampAggregation,
		Interval:              interval,
	})

	if err != nil {

		Logger.Error("Error serializing query", zap.Error(err))

		return nil, err

	}

	receiverChannel := make(chan []byte)

	db.putReceiverChannel(queryId, receiverChannel)

	// Send query
	db.queryChannel <- queryBytes

	// Receive result
	var result Result

	select {

	case <-time.NewTimer(40 * time.Second).C:

		Logger.Info("Query timeout", zap.Uint64("queryId", queryId))

		return nil, ErrQueryTimedOut

	case resultBytes := <-receiverChannel:

		if len(resultBytes) == 0 {

			// Empty result due to closing of receiver channel

			return nil, ErrServerShutdown

		}

		if err = msgpack.Unmarshal(resultBytes, &result); err != nil {

			Logger.Info("Error deserializing query result", zap.Error(err))

			return nil, err
		}

	}

	return parseResponse(result.Data), nil
}

func (db *ReportDBClient) SendPollData(data []byte) {

	defer func() {

		if err := recover(); err != nil {

			Logger.Info("sender channel already closed", zap.Any("error", err))

		}

	}()

	db.pollDataChannel <- data

}

func (db *ReportDBClient) Close() {

	close(db.queryChannel)

	close(db.pollDataChannel)

	db.shutdownChannel <- struct{}{}

	err := db.context.Term()

	if err != nil {

		Logger.Error("Error terminating context", zap.Error(err))

		return
	}

	<-db.shutdownChannel

	db.closeReceivers()

}
