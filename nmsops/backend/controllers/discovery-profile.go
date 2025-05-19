package controllers

import (
	"go.uber.org/zap"
	. "nms-backend/db"
	. "nms-backend/models"
	. "nms-backend/services"
	. "nms-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// discoveryProfileRequest represents the request body for creating and updating a discovery profile
type discoveryProfileRequest struct {
	DeviceIPs            interface{} `json:"device_ips" binding:"required"`
	CredentialProfileIDs []int       `json:"credential_profile_ids" binding:"required"`
	IsCIDR               bool        `json:"is_cidr"`
}

// discoveryProfileResponse represents the response for discovery profile endpoints
type discoveryProfileResponse struct {
	Profiles []DiscoveryProfile `json:"profiles"`
}

type DiscoveryProfileController struct {
	db *ConfigDBClient
}

func NewDiscoveryProfileController(db *ConfigDBClient) *DiscoveryProfileController {
	return &DiscoveryProfileController{db: db}
}

// Create handles POST request to create a new discovery profile
func (discoveryProfileController *DiscoveryProfileController) Create(ctx *gin.Context) {

	defer func() {

		if err := recover(); err != nil {

			Logger.Error("Invalid request body caused panic.", zap.Any("error", err))

			ctx.JSON(400, gin.H{"error": ErrInvalidRequestBody.Error()})

		}

	}()

	var req discoveryProfileRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		Logger.Error("Error parsing request body", zap.Error(err))

		ctx.JSON(400, gin.H{"error": ErrInvalidRequestBody.Error()})

		return

	}

	deviceIps := make([]string, 0)

	if req.IsCIDR {

		if valid := ValidateCIDRIp(req.DeviceIPs.(string)); !valid {

			ctx.JSON(400, gin.H{"error": ErrInvalidCIDRIp.Error()})

			return

		}

		deviceIps = append(deviceIps, GetIpListFromCIDRNetworkIp(req.DeviceIPs.(string))...)

	} else {

		for _, Ip := range req.DeviceIPs.([]interface{}) {

			if valid := ValidateIpAddress(Ip.(string)); !valid {

				ctx.JSON(400, gin.H{"error": ErrInvalidCIDRIp.Error()})

				return

			}

			deviceIps = append(deviceIps, Ip.(string))

		}

	}

	profileID, err := CreateDiscoveryProfile(discoveryProfileController.db, deviceIps, req.CredentialProfileIDs)

	if err != nil {

		Logger.Error("Error creating discovery profile", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to create discovery profile"})

		return

	}

	ctx.JSON(201, gin.H{

		"message": "Discovery profile created successfully",

		"id": profileID,
	})
}

// GetAll handles GET request to fetch all discovery profiles
func (discoveryProfileController *DiscoveryProfileController) GetAll(ctx *gin.Context) {

	profiles, err := GetAllDiscoveryProfiles(discoveryProfileController.db)

	if err != nil {

		Logger.Error("Error querying discovery profiles", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to query discovery profiles"})

		return

	}

	ctx.JSON(200, discoveryProfileResponse{Profiles: profiles})
}

// Update handles PUT request to update an existing discovery profile
func (discoveryProfileController *DiscoveryProfileController) Update(ctx *gin.Context) {

	profileID, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {

		Logger.Error("Error parsing discovery profile id", zap.Error(err))

		ctx.JSON(400, gin.H{"error": "Invalid profile ID"})

		return

	}

	var req discoveryProfileRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		Logger.Error("Error parsing request body", zap.Error(err))

		ctx.JSON(400, gin.H{"error": ErrInvalidRequestBody.Error()})

		return

	}

	deviceIps := make([]string, 0)

	if req.IsCIDR {

		if valid := ValidateCIDRIp(req.DeviceIPs.(string)); !valid {

			ctx.JSON(400, gin.H{"error": ErrInvalidCIDRIp.Error()})

			return

		}

		deviceIps = append(deviceIps, GetIpListFromCIDRNetworkIp(req.DeviceIPs.(string))...)

	} else {

		for _, Ip := range req.DeviceIPs.([]interface{}) {

			if valid := ValidateIpAddress(Ip.(string)); !valid {

				ctx.JSON(400, gin.H{"error": ErrInvalidCIDRIp.Error()})

				return

			}

			deviceIps = append(deviceIps, Ip.(string))

		}

	}

	rowsAffected, err := UpdateDiscoveryProfile(discoveryProfileController.db, profileID, deviceIps, req.CredentialProfileIDs)

	if err != nil {

		Logger.Error("Error updating discovery profile", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to update discovery profile"})

		return

	}

	if rowsAffected == 0 {

		ctx.JSON(404, gin.H{"error": "Discovery profile not found"})

		return

	}

	ctx.JSON(200, gin.H{"message": "Discovery profile updated successfully"})
}

func (discoveryProfileController *DiscoveryProfileController) RunDiscovery(ctx *gin.Context) {

	discoveryProfileID, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {

		Logger.Error("Error parsing discovery profile id", zap.Error(err))

		ctx.JSON(400, gin.H{"error": "Invalid discoveryProfile ID"})

		return

	}

	profile, err := GetDiscoveryProfile(discoveryProfileController.db, discoveryProfileID)

	if err != nil {

		Logger.Error("Error querying discovery profile", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Error querying discovery profile"})

		return
	}

	credentials, err := GetCredentialProfiles(discoveryProfileController.db, profile.CredentialProfileIDs)

	if err != nil {

		Logger.Error("Error querying credential profiles", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Error querying credential profiles"})

		return

	}

	// Run Discovery
	discoveredDevices := Discover(profile.DeviceIPs, credentials)

	err = InsertDiscoveredDevices(discoveryProfileController.db, discoveredDevices)

	if err != nil {

		Logger.Error("Error inserting discovered devices", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to insert discovered devices"})

		return

	}

	ctx.JSON(200, gin.H{"message": "Discovery successfully", "discovered_devices": discoveredDevices, "device_count": len(discoveredDevices)})

}
