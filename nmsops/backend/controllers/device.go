package controllers

import (
	"errors"
	. "nms-backend/db"
	. "nms-backend/models"

	"github.com/gin-gonic/gin"
	"strconv"
)

type DeviceController struct {
	db *ConfigDB
}

func NewDeviceController(db *ConfigDB) *DeviceController {
	return &DeviceController{db: db}
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

// UpdateProvisionStatus handles PUT request to update device provision status
func (deviceController *DeviceController) UpdateProvisionStatus(ctx *gin.Context) {

	ip, err := strconv.ParseUint(ctx.Param("ip"), 10, 32)

	if err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid IP address"})
		return
	}

	var req struct {
		IsProvisioned bool `json:"is_provisioned"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	// First check if the device exists
	checkQuery := `SELECT ip FROM device WHERE ip = $1`
	var existingIP uint32
	err = deviceController.db.QueryRow(checkQuery, ip).Scan(&existingIP)
	if err != nil {
		ctx.JSON(404, gin.H{"error": "Device not found"})
		return
	}

	query := `UPDATE device SET is_provisioned = $1 WHERE ip = $2`
	result, err := deviceController.db.Exec(query, req.IsProvisioned, ip)
	if err != nil {
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

	ctx.JSON(200, gin.H{"message": "Device provision status updated successfully"})
}

// UpdateProvisionStatus handles PUT request to update device provision status
func (deviceController *DeviceController) UpdateProvisionStatusV2(ctx *gin.Context) {

	var req DeviceProvisionUpdateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid request body"})

		return

	}

	query := `UPDATE device SET is_provisioned = NOT is_provisioned WHERE ip = ANY($1)`

	result, err := deviceController.db.Exec(query, req.ProvisionUpdateIps)

	if err != nil {

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

	ctx.JSON(200, gin.H{"message": "Device provision status updated successfully"})
}

func InsertDiscoveredDevices(db *ConfigDB, devices []Device) error {
	for _, device := range devices {

		query := `
		INSERT INTO device (ip, credential_id, is_provisioned)
		VALUES ($1, $2, $3)
		RETURNING ip`

		var ip uint32

		err := db.QueryRow(query, device.IP, device.CredentialID, device.IsProvisioned).Scan(&ip)

		if err != nil {

			return errors.New("failed to insert discovered device")

		}

	}

	return nil
}
