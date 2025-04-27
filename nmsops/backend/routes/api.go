package routes

import (
	. "nms-backend/controllers"
	. "nms-backend/db"
	. "nms-backend/reportdb"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, reportDB *ReportDBClient, mainDB *ConfigDB) {

	api := router.Group("/api")

	queryController := NewQueryController(reportDB)

	deviceController := NewDeviceController(mainDB)

	credentialProfileController := NewCredentialProfileController(mainDB)

	discoveryProfileController := NewDiscoveryProfileController(mainDB)

	// Use POST for histogram queries with body parameters
	api.POST("/query", queryController.HandleQuery)

	// Device endpoints
	api.GET("/devices", deviceController.GetAllDevices)

	api.PUT("/devices/:ip/provision", deviceController.UpdateProvisionStatus)

	// Credential Profile endpoints
	api.GET("/credential-profiles", credentialProfileController.GetCredentialProfiles)

	api.POST("/credential-profiles", credentialProfileController.CreateCredentialProfile)

	api.PUT("/credential-profiles/:id", credentialProfileController.UpdateCredentialProfile)

	api.DELETE("/credential-profiles/:id", credentialProfileController.DeleteCredentialProfile)

	// Discovery Profile endpoints
	api.GET("/discovery-profiles", discoveryProfileController.GetDiscoveryProfiles)

	api.GET("/discovery-profiles/:id/run-discovery", discoveryProfileController.RunDiscovery)

	api.POST("/discovery-profiles", discoveryProfileController.CreateDiscoveryProfile)

	api.PUT("/discovery-profiles/:id", discoveryProfileController.UpdateDiscoveryProfile)

	// future:
	// api.GET("/devices", ...)
	// api.POST("/alerts", ...)
}
