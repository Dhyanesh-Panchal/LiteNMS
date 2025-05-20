package controllers

import (
	. "nms-backend/db"
	. "nms-backend/models"
	. "nms-backend/services"
	. "nms-backend/utils"

	"github.com/gin-gonic/gin"
)

type deviceResponse struct {
	Devices []Device `json:"devices"`
}

type deviceProvisionUpdateRequest struct {
	ProvisionUpdateIps []string `json:"provision_update_ips"`
}

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

// GetAll handles the GET request to fetch all devices
func (deviceController *DeviceController) GetAll(ctx *gin.Context) {

	devices, err := GetAllDevices(deviceController.db)

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to query devices"})

		return

	}

	response := deviceResponse{

		Devices: devices,
	}

	ctx.JSON(200, response)
}

// UpdateProvisionStatus handles PUT request to update devices provision status
func (deviceController *DeviceController) UpdateProvisionStatus(ctx *gin.Context) {

	var req deviceProvisionUpdateRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid request body"})

		return

	}

	// Validate IP addresses
	for _, ip := range req.ProvisionUpdateIps {

		if valid := ValidateIpAddress(ip); !valid {

			ctx.JSON(400, gin.H{"error": "Invalid IP address"})

			return

		}

	}

	rowsAffected, err := UpdateDeviceProvisionStatus(deviceController.db, req.ProvisionUpdateIps)

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to update device provision status"})

		return

	}

	err = deviceController.provisioningPublisher.SendUpdate(req.ProvisionUpdateIps, "")

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to publish device provision status to polling engine"})

		return

	}

	ctx.JSON(200, gin.H{"message": "Device provision status updated successfully", "provision_count": rowsAffected})
}
