package controllers

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	. "nms-backend/db"
	. "nms-backend/models"
	. "nms-backend/services"

	"github.com/gin-gonic/gin"
)

type DeviceController struct {
	db *ConfigDBClient

	provisioningPublisher *ProvisioningPublisher
}

func NewDeviceController(db *ConfigDBClient, provisioningPublisher *ProvisioningPublisher) *DeviceController {

	return &DeviceController{

		db: db,

		provisioningPublisher: provisioningPublisher,
	}

}

// GetAllDevices handles the GET request to fetch all devices
func (deviceController *DeviceController) GetAllDevices(ctx *gin.Context) {

	// Query the database for all devices
	query := `SELECT ip, credential_id, is_provisioned FROM device`

	rows, err := deviceController.db.Query(query)

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to query devices"})

		return

	}

	defer rows.Close()

	var devices []Device

	for rows.Next() {

		var device Device

		if err := rows.Scan(&device.IP, &device.CredentialID, &device.IsProvisioned); err != nil {

			ctx.JSON(500, gin.H{"error": "Failed to scan device row"})

			return

		}

		devices = append(devices, device)

	}

	if err := rows.Err(); err != nil {

		ctx.JSON(500, gin.H{"error": "Error iterating over device rows"})

		return

	}

	response := DeviceResponse{

		Devices: devices,
	}

	ctx.JSON(200, response)
}

// UpdateProvisionStatus handles PUT request to update devices provision status
func (deviceController *DeviceController) UpdateProvisionStatus(ctx *gin.Context) {

	var req DeviceProvisionUpdateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid request body"})

		return

	}

	query := `UPDATE device SET is_provisioned = NOT is_provisioned WHERE ip = ANY($1)`

	result, err := deviceController.db.Exec(query, pq.Array(req.ProvisionUpdateIps))

	if err != nil {

		fmt.Println(err)

		ctx.JSON(500, gin.H{"error": "Failed to update device provision status"})

		return

	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to get rows affected"})

		return

	}

	if rowsAffected == 0 {

		ctx.JSON(404, gin.H{"error": "Device not found"})

		return

	}

	err = deviceController.provisioningPublisher.SendUpdate(req.ProvisionUpdateIps, "")

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to publish device provision status to polling engine"})

		return

	}

	ctx.JSON(200, gin.H{"message": "Device provision status updated successfully", "provision_count": rowsAffected})
}

func InsertDiscoveredDevices(db *ConfigDBClient, devices []Device) error {
	for _, device := range devices {

		query := `
		INSERT INTO device (ip, credential_id, is_provisioned)
		VALUES ($1, $2, $3)
		RETURNING ip`

		var ip uint32

		err := db.QueryRow(query, device.IP, device.CredentialID, device.IsProvisioned).Scan(&ip)

		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {

				return nil

			}

			return errors.New("failed to insert discovered device")

		}

	}

	return nil
}
