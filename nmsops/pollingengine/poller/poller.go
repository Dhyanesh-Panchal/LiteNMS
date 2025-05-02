package poller

import (
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	. "poller/utils"
	"strconv"
	"strings"
	"sync"
)

type PolledDataPoint struct {
	Timestamp uint32 `json:"timestamp" msgpack:"timestamp"`

	CounterId uint16 `json:"counter_id" msgpack:"counter_id"`

	ObjectId uint32 `json:"object_id" msgpack:"object_id"`

	Value interface{} `json:"value" msgpack:"value"`
}

type PollJob struct {
	Timestamp uint32

	DeviceIP string

	DeviceConfig *ssh.ClientConfig

	DevicePort string

	CounterId uint16
}

var CounterCommand = map[uint16]string{
	1: "free -m | awk 'NR==2 {print $3}'",
	2: "top -bn 1 | awk 'NR==3 {print $2}'",
	3: "whoami",
}

func Poller(pollJobChannel <-chan PollJob, pollResultChannel chan<- PolledDataPoint, shutdownWaitGroup *sync.WaitGroup) {

	defer shutdownWaitGroup.Done()

	for job := range pollJobChannel {

		resp, err := poll(job.DeviceConfig, job.DeviceIP, job.DevicePort, CounterCommand[job.CounterId])

		if err != nil {

			continue

		}

		var value interface{}

		switch CounterConfig[job.CounterId]["dataType"] {

		case "int", "int32", "int64", "uint", "uint32", "uint64":
			value, _ = strconv.Atoi(resp)

		case "float32", "float64":
			value, _ = strconv.ParseFloat(resp, 64)

		case "string":
			value = resp

		}

		//value := resp

		dataPoint := PolledDataPoint{

			Timestamp: job.Timestamp,

			ObjectId: ConvertIpToNumeric(job.DeviceIP),

			CounterId: job.CounterId,

			Value: value,
		}

		pollResultChannel <- dataPoint

		Logger.Info("Poll success for", zap.String("ObjectId", job.DeviceIP), zap.Any("DataPoint", dataPoint))

	}

	Logger.Info("Poller exiting")

}

func poll(config *ssh.ClientConfig, deviceIp, port, cmd string) (string, error) {

	client, err := ssh.Dial("tcp", deviceIp+":"+port, config)

	if err != nil {

		Logger.Info("Error dialing ssh connection", zap.String("Device IP", deviceIp), zap.String("port", port), zap.Error(err))

		return "", err

	}

	defer client.Close()

	session, err := client.NewSession()

	if err != nil {

		Logger.Error("Failed to create session:", zap.Error(err))

		return "", err

	}

	defer session.Close()

	resp, err := session.CombinedOutput(cmd)

	if err != nil {

		Logger.Error("Failed to execute command:", zap.Error(err))

		return "", err
	}

	return strings.TrimRight(string(resp), "\n"), nil

}
