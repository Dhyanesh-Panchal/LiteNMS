package containers

import (
	"database/sql"
	"github.com/lib/pq"
	"go.uber.org/zap"
	. "poller/utils"
	"strconv"
	"sync"
)

const (
	allDevicesQuery = `
		SELECT d.ip, c.hostname, c.password, c.port  FROM device d
		JOIN credential_profiles c ON d.credential_id = c.credential_profile_id
		WHERE d.is_provisioned = TRUE;
`

	specificDevicesQuery = `
		SELECT d.ip, c.hostname, c.password, c.port, d.is_provisioned  FROM device d
		JOIN credential_profiles c ON d.credential_id = c.credential_profile_id
		WHERE d.ip = ANY($1);
`
)

type PollJob struct {
	Timestamp uint32

	DeviceIP string

	Hostname string

	Password string

	Port string

	CounterIds []uint16
}

type DeviceList struct {
	deviceConfig map[string][3]string

	db *sql.DB

	lock sync.RWMutex
}

func NewDeviceList() (*DeviceList, error) {

	db, err := sql.Open("postgres", GetConfigDBConnectionString())

	if err != nil {

		Logger.Error("unable to connect to configDB", zap.Error(err))

		return nil, err

	}

	// Get the all provisioned deviceConfig from the configDB

	rows, err := db.Query(allDevicesQuery)

	if err != nil {

		Logger.Error("failed to query deviceConfig", zap.Error(err))

		return nil, err

	}

	defer func(rows *sql.Rows) {

		err := rows.Close()

		if err != nil {

			Logger.Error("failed to close rows stream", zap.Error(err))

		}

	}(rows)

	// device config as map of ip to [hostname, password, port]
	devices := make(map[string][3]string)

	for rows.Next() {

		var ip, hostname, password string

		var port int

		if err := rows.Scan(&ip, &hostname, &password, &port); err != nil {

			Logger.Error("Failed to scan device", zap.Error(err))

			continue

		}

		devices[ip] = [3]string{hostname, password, strconv.Itoa(port)}

	}

	return &DeviceList{

		deviceConfig: devices,

		db: db,
	}, nil

}

func (list *DeviceList) UpdateProvisionedDeviceList(statusUpdateIps []string) {

	list.lock.Lock()

	defer list.lock.Unlock()

	rows, err := list.db.Query(specificDevicesQuery, pq.Array(statusUpdateIps))

	if err != nil {

		Logger.Error("Failed to query deviceConfig", zap.Error(err))

		return

	}

	defer func(rows *sql.Rows) {

		err := rows.Close()

		if err != nil {

			Logger.Error("Failed to close rows", zap.Error(err))

		}

	}(rows)

	for rows.Next() {

		var ip, hostname, password string

		var port int

		var isProvisioned bool

		if err := rows.Scan(&ip, &hostname, &password, &port, &isProvisioned); err != nil {

			Logger.Error("Failed to scan device", zap.Error(err))

			continue

		}

		if isProvisioned {

			// New Device provisioned

			list.deviceConfig[ip] = [3]string{hostname, password, strconv.Itoa(port)}

		} else {

			// Device Unprovisioned

			Logger.Info("Unprovisioning device", zap.String("IP:", ip))

			delete(list.deviceConfig, ip)

		}

	}

}

func (list *DeviceList) PreparePollJobs(timestamp uint32, qualifiedCounterIds []uint16) []PollJob {

	list.lock.RLock()

	defer list.lock.RUnlock()

	var pollJobs []PollJob

	for ip, config := range list.deviceConfig {

		pollJobs = append(pollJobs, PollJob{
			Timestamp:  timestamp,
			DeviceIP:   ip,
			Hostname:   config[0],
			Password:   config[1],
			Port:       config[2],
			CounterIds: qualifiedCounterIds,
		})

	}

	return pollJobs

}

func (list *DeviceList) Close() {

	list.lock.Lock()

	defer list.lock.Unlock()

	clear(list.deviceConfig)

	list.db.Close()

}
