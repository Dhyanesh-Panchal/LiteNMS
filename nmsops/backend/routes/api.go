package routes

import (
	"github.com/gin-gonic/gin"
	. "nms-backend/controllers"
	. "nms-backend/db"
	. "nms-backend/services"
)

func SetupRoutes(router *gin.Engine, reportDB *ReportDbClient, mainDB *ConfigDB, provisioningPublisher *ProvisioningPublisher) {

	api := router.Group("/api")

	queryController := NewQueryController(reportDB)

	deviceController := NewDeviceController(mainDB, provisioningPublisher)

	credentialProfileController := NewCredentialProfileController(mainDB)

	discoveryProfileController := NewDiscoveryProfileController(mainDB)

	api.POST("/query", queryController.HandleQuery)

	// Device endpoints
	api.GET("/devices", deviceController.GetAllDevices)

	api.PUT("/devices/update-provisioning", deviceController.UpdateProvisionStatusV2)

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
