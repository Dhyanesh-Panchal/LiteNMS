package controllers

import (
	. "nms-backend/db"
	. "nms-backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

const CredentialProfileNotFound = "Credential profile not found"

type CredentialProfileController struct {
	db *ConfigDBClient
}

func NewCredentialProfileController(db *ConfigDBClient) *CredentialProfileController {

	return &CredentialProfileController{db: db}
}

// GetCredentialProfiles handles GET request to fetch all credential profiles
func (credentialProfileController *CredentialProfileController) GetCredentialProfiles(ctx *gin.Context) {

	query := `SELECT credential_profile_id, hostname, password, port FROM credential_profiles`

	rows, err := credentialProfileController.db.Query(query)

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to query credential profiles"})

		return

	}

	defer rows.Close()

	var profiles []CredentialProfile

	for rows.Next() {

		var profile CredentialProfile

		if err := rows.Scan(&profile.ID, &profile.Hostname, &profile.Password, &profile.Port); err != nil {

			ctx.JSON(500, gin.H{"error": "Failed to scan credential profile"})

			return

		}

		profiles = append(profiles, profile)

	}

	if err := rows.Err(); err != nil {

		ctx.JSON(500, gin.H{"error": "Error iterating over credential profiles"})

		return

	}

	ctx.JSON(200, CredentialProfileResponse{Profiles: profiles})

}

// CreateCredentialProfile handles POST request to create a new credential profile
func (credentialProfileController *CredentialProfileController) CreateCredentialProfile(ctx *gin.Context) {

	var req CredentialProfileRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid request body"})

		return

	}

	query := `
		INSERT INTO credential_profiles (hostname, password, port)
		VALUES ($1, $2, $3)
		RETURNING credential_profile_id`

	var profileID int

	err := credentialProfileController.db.QueryRow(query, req.Hostname, req.Password, req.Port).Scan(&profileID)

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to create credential profile"})

		return

	}

	ctx.JSON(201, gin.H{

		"message": "Credential profile created successfully",

		"id": profileID,
	})

}

// UpdateCredentialProfile handles PUT request to update an existing credential profile
func (credentialProfileController *CredentialProfileController) UpdateCredentialProfile(ctx *gin.Context) {

	profileID, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid profile ID"})

		return

	}

	var req CredentialProfileRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid request body"})

		return

	}

	query := `
		UPDATE credential_profiles
		SET hostname = $1, password = $2, port = $3
		WHERE credential_profile_id = $4`

	result, err := credentialProfileController.db.Exec(query, req.Hostname, req.Password, req.Port, profileID)

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to update credential profile"})

		return

	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to get rows affected"})

		return

	}

	if rowsAffected == 0 {

		ctx.JSON(404, gin.H{"error": CredentialProfileNotFound})

		return

	}

	ctx.JSON(200, gin.H{"message": "Credential profile updated successfully"})

}

// DeleteCredentialProfile handles DELETE request to remove a credential profile
func (credentialProfileController *CredentialProfileController) DeleteCredentialProfile(ctx *gin.Context) {

	profileID, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid profile ID"})

		return

	}

	// First check if the profile exists
	checkQuery := `SELECT credential_profile_id FROM credential_profiles WHERE credential_profile_id = $1`

	var existingID int

	err = credentialProfileController.db.QueryRow(checkQuery, profileID).Scan(&existingID)

	if err != nil {

		ctx.JSON(404, gin.H{"error": CredentialProfileNotFound})

		return

	}

	query := `DELETE FROM credential_profiles WHERE credential_profile_id = $1`

	result, err := credentialProfileController.db.Exec(query, profileID)

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to delete credential profile"})

		return

	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to get rows affected"})

		return

	}

	if rowsAffected == 0 {

		ctx.JSON(404, gin.H{"error": CredentialProfileNotFound})

		return

	}

	ctx.JSON(200, gin.H{"message": "Credential profile deleted successfully"})
}
