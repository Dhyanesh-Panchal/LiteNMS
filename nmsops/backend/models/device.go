package models

import (
	"errors"
	"github.com/lib/pq"
	. "nms-backend/db"
)

// Device represents a network device in the system
type Device struct {
	IP string `json:"ip" db:"ip"`

	CredentialID int `json:"credential_id" db:"credential_id"`

	IsProvisioned bool `json:"is_provisioned" db:"is_provisioned"`
}

func InsertDiscoveredDevices(db *ConfigDBClient, devices []Device) error {
	for _, device := range devices {

		query := `
		INSERT INTO device (ip, credential_id, is_provisioned)
		VALUES ($1, $2, $3)
		RETURNING ip`

		var ip string

		err := db.QueryRow(query, device.IP, device.CredentialID, device.IsProvisioned).Scan(&ip)

		if err != nil {

			var pqErr *pq.Error

			// device already exist
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {

				continue

			}

			return err

		}

	}

	return nil
}

func GetAllDevices(db *ConfigDBClient) ([]Device, error) {
	query := `SELECT ip, credential_id, is_provisioned FROM device`

	rows, err := db.Query(query)

	if err != nil {

		return nil, err

	}

	defer rows.Close()

	var devices []Device

	for rows.Next() {

		var device Device

		if err = rows.Scan(&device.IP, &device.CredentialID, &device.IsProvisioned); err != nil {

			return nil, err

		}

		devices = append(devices, device)

	}

	return devices, nil
}

func UpdateDeviceProvisionStatus(db *ConfigDBClient, ips []string) (int64, error) {
	query := `UPDATE device SET is_provisioned = NOT is_provisioned WHERE ip = ANY($1)`

	result, err := db.Exec(query, pq.Array(ips))

	if err != nil {

		return -1, err

	}

	return result.RowsAffected()
}
