package controllers

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
	. "nms-backend/db"
	. "nms-backend/models"
	. "nms-backend/services"
	. "nms-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type DiscoveryProfileController struct {
	db *ConfigDBClient
}

func NewDiscoveryProfileController(db *ConfigDBClient) *DiscoveryProfileController {
	return &DiscoveryProfileController{db: db}
}

// GetDiscoveryProfiles handles GET request to fetch all discovery profiles
func (discoveryProfileController *DiscoveryProfileController) GetDiscoveryProfiles(ctx *gin.Context) {
	query := `SELECT discovery_profile_id, device_ips, credential_profiles FROM discovery_profile`

	rows, err := discoveryProfileController.db.Query(query)

	if err != nil {

		Logger.Error("Error querying discovery profiles:", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to query discovery profiles"})

		return

	}

	defer rows.Close()

	var profiles []DiscoveryProfile

	for rows.Next() {

		profile, err := parseNextDiscoveryProfileRow(rows)

		if err != nil {

			Logger.Error("Error parsing discovery profile row", zap.Error(err))

			ctx.JSON(500, gin.H{"error": "Failed to scan discovery profile"})

		}

		profiles = append(profiles, profile)
	}

	ctx.JSON(200, DiscoveryProfileResponse{Profiles: profiles})
}

func parseNextDiscoveryProfileRow(rows *sql.Rows) (DiscoveryProfile, error) {
	var profile DiscoveryProfile

	var deviceIPs []int64

	var credentialProfiles []int64

	if err := rows.Scan(&profile.ID, pq.Array(&deviceIPs), pq.Array(&credentialProfiles)); err != nil {

		return profile, errors.New("failed to scan discovery profile")

	}

	profile.DeviceIPs = make([]uint32, len(deviceIPs))

	profile.CredentialProfileIDs = make([]int, len(credentialProfiles))

	for i := range len(deviceIPs) {

		profile.DeviceIPs[i] = uint32(deviceIPs[i])

	}
	for i := range len(credentialProfiles) {

		profile.CredentialProfileIDs[i] = int(credentialProfiles[i])

	}

	return profile, nil
}

// CreateDiscoveryProfile handles POST request to create a new discovery profile
func (discoveryProfileController *DiscoveryProfileController) CreateDiscoveryProfile(ctx *gin.Context) {

	var req DiscoveryProfileRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		Logger.Error("Error parsing request body", zap.Error(err))

		ctx.JSON(400, gin.H{"error": "Invalid request body"})

		return

	}

	query := `
		INSERT INTO discovery_profile (device_ips, credential_profiles)
		VALUES ($1, $2)
		RETURNING discovery_profile_id`

	var profileID int

	err := discoveryProfileController.db.QueryRow(query, pq.Array(req.DeviceIPs), pq.Array(req.CredentialProfileIDs)).Scan(&profileID)

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

// UpdateDiscoveryProfile handles PUT request to update an existing discovery profile
func (discoveryProfileController *DiscoveryProfileController) UpdateDiscoveryProfile(ctx *gin.Context) {

	profileID, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {

		Logger.Error("Error parsing discovery profile id", zap.Error(err))

		ctx.JSON(400, gin.H{"error": "Invalid profile ID"})

		return

	}

	var req DiscoveryProfileRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		Logger.Error("Error parsing request body", zap.Error(err))

		ctx.JSON(400, gin.H{"error": "Invalid request body"})

		return

	}

	query := `
		UPDATE discovery_profile
		SET device_ips = $1, credential_profiles = $2
		WHERE discovery_profile_id = $3`

	result, err := discoveryProfileController.db.Exec(query, pq.Array(req.DeviceIPs), pq.Array(req.CredentialProfileIDs), profileID)

	if err != nil {

		Logger.Error("Error updating discovery profile", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to update discovery profile"})

		return

	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {

		Logger.Error("Error getting rows affected", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to get rows affected"})

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

	query := `SELECT discovery_profile_id, device_ips, credential_profiles FROM discovery_profile WHERE discovery_profile_id = $1`

	rows, err := discoveryProfileController.db.Query(query, discoveryProfileID)

	if err != nil {

		Logger.Error("Error querying discovery profiles", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to run discovery discoveryProfile"})

	}

	defer rows.Close()

	if result := rows.Next(); !result {

		ctx.JSON(404, gin.H{"error": "Discovery discoveryProfile not found"})

		return
	}

	discoveryProfile, err := parseNextDiscoveryProfileRow(rows)

	if err != nil {

		Logger.Error("Error parsing discovery profile row", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to scan discovery profile"})

		return
	}

	query = `SELECT credential_profile_id, hostname, password, port FROM credential_profiles WHERE credential_profile_id = ANY($1) `

	rows, err = discoveryProfileController.db.Query(query, pq.Array(discoveryProfile.CredentialProfileIDs))

	if err != nil {

		Logger.Error("Error querying Credential profiles", zap.Error(err))

		ctx.JSON(500, gin.H{"error": "Failed to retrieve credential profiles"})

	}

	defer rows.Close()

	var credentials []CredentialProfile

	for rows.Next() {

		var profile CredentialProfile

		if err := rows.Scan(&profile.ID, &profile.Hostname, &profile.Password, &profile.Port); err != nil {

			ctx.JSON(500, gin.H{"error": "Failed to scan credential discoveryProfile"})

			return

		}

		credentials = append(credentials, profile)

	}

	if err := rows.Err(); err != nil {

		ctx.JSON(500, gin.H{"error": "Error iterating over credential profiles"})

		return

	}

	// Run Discovery

	discoveredDevices := RunDiscovery(discoveryProfile.DeviceIPs, credentials)

	err = InsertDiscoveredDevices(discoveryProfileController.db, discoveredDevices)

	if err != nil {

		Logger.Error("Error inserting discovered devices", zap.Error(err))
		
		ctx.JSON(500, gin.H{"error": "Failed to insert discovered devices"})

	}

	ctx.JSON(200, gin.H{"message": "Discovery successfully", "discovered_devices": discoveredDevices, "device_count": len(discoveredDevices)})

}
