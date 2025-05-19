package router

import (
	"github.com/gin-gonic/gin"
	. "nms-backend/controllers"
	. "nms-backend/db"
	. "nms-backend/services"
)

func SetupRoutes(router *gin.Engine, reportDB *ReportDBClient, configDB *ConfigDBClient, provisioningPublisher *ProvisioningPublisher) {

	api := router.Group("/api")

	queryController := NewQueryController(reportDB)

	deviceController := NewDeviceController(configDB, provisioningPublisher)

	credentialProfileController := NewCredentialProfileController(configDB)

	discoveryProfileController := NewDiscoveryProfileController(configDB)

	// Credential Profile endpoints
	api.POST("/credential-profiles", credentialProfileController.Create)

	api.GET("/credential-profiles", credentialProfileController.GetAll)

	api.PUT("/credential-profiles/:id", credentialProfileController.Update)

	// Discovery Profile endpoints
	api.POST("/discovery-profiles", discoveryProfileController.Create)

	api.GET("/discovery-profiles", discoveryProfileController.GetAll)

	api.PUT("/discovery-profiles/:id", discoveryProfileController.Update)

	api.GET("/discovery-profiles/:id", discoveryProfileController.RunDiscovery)

	// Device endpoints
	api.GET("/devices", deviceController.GetAll)

	api.PUT("/devices/update-provisioning", deviceController.UpdateProvisionStatus)

	// Query endpoints
	api.POST("/query", queryController.HandleQuery)

}
