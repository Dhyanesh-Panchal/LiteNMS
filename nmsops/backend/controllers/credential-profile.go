package controllers

import (
	"errors"
	"github.com/lib/pq"
	"go.uber.org/zap"
	. "nms-backend/db"
	. "nms-backend/models"
	. "nms-backend/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// credentialProfileRequest represents the request body for creating and updating a credential profile
type credentialProfileRequest struct {
	Hostname string `json:"hostname" binding:"required"`

	Password string `json:"password" binding:"required"`

	Port uint16 `json:"port" binding:"required"`
}

// credentialProfileResponse represents the response for credential profile endpoints
type credentialProfileResponse struct {
	Profiles []CredentialProfile `json:"profiles"`
}

type CredentialProfileController struct {
	db *ConfigDBClient
}

func NewCredentialProfileController(db *ConfigDBClient) *CredentialProfileController {

	return &CredentialProfileController{db: db}
}

// Create handles POST request to create a new credential profile
func (credentialProfileController *CredentialProfileController) Create(ctx *gin.Context) {

	var req credentialProfileRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid request body"})

		return

	}

	// Validate port

	if req.Port == 0 {

		ctx.JSON(400, gin.H{"error": "0 port is not valid"})

		return

	}

	profileID, err := CreateCredentialProfile(credentialProfileController.db, req.Hostname, req.Password, req.Port)

	if err != nil {

		var pqErr *pq.Error

		// Check for duplicate key error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {

			ctx.JSON(400, gin.H{"error": "Credential profile already exists"})

			return

		} else {

			Logger.Error("Error creating credential profile", zap.Error(err))

			ctx.JSON(500, gin.H{"error": "Failed to create credential profile"})

			return

		}

	}

	ctx.JSON(201, gin.H{

		"message": "Credential profile created successfully",

		"id": profileID,
	})

}

// GetAll handles GET request to fetch all credential profiles
func (credentialProfileController *CredentialProfileController) GetAll(ctx *gin.Context) {

	profiles, err := GetAllCredentialProfiles(credentialProfileController.db)

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to query credential profiles"})

		return

	}

	ctx.JSON(200, credentialProfileResponse{Profiles: profiles})

}

// Update handles PUT request to update an existing credential profile
func (credentialProfileController *CredentialProfileController) Update(ctx *gin.Context) {

	profileID, err := strconv.Atoi(ctx.Param("id"))

	if err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid profile ID"})

		return

	}

	var req credentialProfileRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {

		ctx.JSON(400, gin.H{"error": "Invalid request body"})

		return

	}

	// Validate port
	if req.Port == 0 {

		ctx.JSON(400, gin.H{"error": "0 port is not valid"})

		return

	}

	rowsAffected, err := UpdateCredentialProfile(credentialProfileController.db, profileID, req.Hostname, req.Password, req.Port)

	if err != nil {

		ctx.JSON(500, gin.H{"error": "Failed to update credential profile"})

		return

	}

	if rowsAffected == 0 {

		ctx.JSON(404, gin.H{"error": "Credential profile not found"})

		return

	}

	ctx.JSON(200, gin.H{"message": "Credential profile updated successfully"})

}
