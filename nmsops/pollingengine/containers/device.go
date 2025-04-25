package containers

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
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

type DeviceList struct {
	deviceConfig map[uint32]*ssh.ClientConfig

	devicePort map[uint32]string

	db *pgxpool.Pool

	globalContext context.Context

	lock sync.RWMutex
}

func NewDeviceList(globalContext context.Context) *DeviceList {

	connStr := fmt.Sprintf(

		"postgres://%s:%s@%s:%s/%s",

		ConfigDBUser,

		ConfigDBPassword,

		ConfigDBHost,

		ConfigDBPort,

		ConfigDBName,
	)

	db, err := pgxpool.New(globalContext, connStr)

	if err != nil {

		Logger.Error("Unable to connect to configDB", zap.Error(err))

	}

	// Get the provisioned deviceConfig from the configDB

	rows, err := db.Query(globalContext, allDevicesQuery)

	if err != nil {

		Logger.Error("Failed to query deviceConfig", zap.Error(err))

	}

	defer rows.Close()

	// Create ssh clients for deviceConfig and save it to map

	devices := make(map[uint32]*ssh.ClientConfig)

	ports := make(map[uint32]string)

	for rows.Next() {

		var ip uint32

		var hostname, password string

		var port int

		if err := rows.Scan(&ip, &hostname, &password, &port); err != nil {

			Logger.Error("Failed to scan device", zap.Error(err))

			continue

		}

		devices[ip] = &ssh.ClientConfig{

			User: hostname,

			Auth: []ssh.AuthMethod{

				ssh.Password(password),
			},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		ports[ip] = strconv.Itoa(port)

	}

	return &DeviceList{

		deviceConfig: devices,

		devicePort: ports,

		globalContext: globalContext,

		db: db,
	}

}

func (list *DeviceList) UpdateProvisionedDeviceList(statusUpdateIps []uint32) {

	list.lock.Lock()

	defer list.lock.Unlock()

	rows, err := list.db.Query(list.globalContext, specificDevicesQuery, statusUpdateIps)

	if err != nil {

		Logger.Error("Failed to query deviceConfig", zap.Error(err))

	}

	defer rows.Close()

	for rows.Next() {

		var ip uint32

		var hostname, password string

		var port int

		var isProvisioned bool

		if err := rows.Scan(&ip, &hostname, &password, &port, &isProvisioned); err != nil {

			Logger.Error("Failed to scan device", zap.Error(err))

			continue

		}

		if isProvisioned {

			// New Device provisioned

			list.deviceConfig[ip] = &ssh.ClientConfig{

				User: hostname,

				Auth: []ssh.AuthMethod{

					ssh.Password(password),
				},

				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}

		} else {

			// Device Unprovisioned

			Logger.Info("Unprovisioning device", zap.Uint32("IP:", ip))

			delete(list.deviceConfig, ip)

		}

	}

}

func (list *DeviceList) GetDevices() (map[uint32]*ssh.ClientConfig, map[uint32]string) {

	list.lock.RLock()

	defer list.lock.RUnlock()

	return list.deviceConfig, list.devicePort

}

func (list *DeviceList) Close() {

	list.lock.Lock()

	defer list.lock.Unlock()

	clear(list.deviceConfig)

	list.db.Close()

}
