package poll

import (
	"context"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
	. "poller/containers"
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

var CounterCommand = map[uint16]string{
	1: "free -m | awk 'NR==2 {print $3}'",
	2: "top -bn 1 | awk 'NR==3 {print $2}'",
	3: "whoami",
}

func InitPollers(pollJobChannel <-chan PollJob, pollResultChannel chan<- PolledDataPoint, globalShutdownChannel <-chan struct{}, globalShutdownWaitGroup *sync.WaitGroup) {

	defer globalShutdownWaitGroup.Done()

	pollerShutdownContext, cancel := context.WithCancel(context.Background())

	var pollerShutdownWaitGroup sync.WaitGroup

	pollerShutdownWaitGroup.Add(PollWorkers)

	for range PollWorkers {

		go Poller(pollJobChannel, pollResultChannel, pollerShutdownContext, &pollerShutdownWaitGroup)

	}

	<-globalShutdownChannel

	cancel()

	pollerShutdownWaitGroup.Wait()

	Logger.Debug("All Pollers exited")

	close(pollResultChannel)

}

func Poller(pollJobChannel <-chan PollJob, pollResultChannel chan<- PolledDataPoint, pollerShutdownContext context.Context, pollerShutdownWaitGroup *sync.WaitGroup) {

	defer pollerShutdownWaitGroup.Done()

	for {

		select {

		case <-pollerShutdownContext.Done():

			Logger.Debug("Poller Exiting")

			return

		default:
			job := <-pollJobChannel

			if job.Timestamp == 0 {

				// Channel closed

				continue

			}

			// Prepare the command
			var command string

			for _, counterId := range job.CounterIds {

				command += CounterCommand[counterId] + ";echo " + CommandDelimiter + ";"

			}

			// Poll
			resp, err := pollDevice(job.DeviceIP, job.Hostname, job.Password, job.Port, command)

			if err != nil {

				continue

			}

			for index, counterId := range job.CounterIds {

				var value interface{}

				switch CounterConfig[counterId]["dataType"] {

				case "int", "int32", "int64", "uint", "uint32", "uint64":

					value, err = strconv.Atoi(resp[index])

					if err != nil {

						Logger.Error("error converting string to int", zap.String("value", resp[index]), zap.Uint16("counterId", counterId), zap.Error(err))

						continue

					}

				case "float32", "float64":

					value, err = strconv.ParseFloat(resp[index], 64)

					if err != nil {

						Logger.Error("error converting string to float", zap.String("value", resp[index]), zap.Uint16("counterId", counterId), zap.Error(err))

						continue

					}

				case "string":

					value = resp[index]

				}

				dataPoint := PolledDataPoint{

					Timestamp: job.Timestamp,

					ObjectId: ConvertIpToNumeric(job.DeviceIP),

					CounterId: counterId,

					Value: value,
				}

				pollResultChannel <- dataPoint

				Logger.Debug("pollDevice success for", zap.String("ObjectId", job.DeviceIP), zap.Any("DataPoint", dataPoint))
			}

		}
	}

}

func pollDevice(deviceIp, hostname, password, port, cmd string) ([]string, error) {

	config := ssh.ClientConfig{

		User: hostname,

		Auth: []ssh.AuthMethod{

			ssh.Password(password),
		},

		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", deviceIp+":"+port, &config)

	if err != nil {

		Logger.Info("error dialing ssh connection", zap.String("Device IP", deviceIp), zap.String("port", port), zap.Error(err))

		return nil, err

	}

	defer func(client *ssh.Client) {

		err := client.Close()

		if err != nil {

			Logger.Warn("error closing the ssh client", zap.String("Device IP", deviceIp), zap.String("port", port), zap.Error(err))

		}

	}(client)

	session, err := client.NewSession()

	if err != nil {

		Logger.Error("failed to create session:", zap.Error(err))

		return nil, err

	}

	defer func(session *ssh.Session) {

		_ = session.Close()

	}(session)

	resp, err := session.CombinedOutput(cmd)

	if err != nil {

		Logger.Error("failed to execute command:", zap.Error(err))

		return nil, err
	}

	return strings.Split(string(resp), "\n"+CommandDelimiter+"\n"), nil

}
