package routes

import (
	. "nms-backend/controllers"
	. "nms-backend/db"
	. "nms-backend/models"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, reportDB ReportDB, mainDB *ConfigDB) {

	api := router.Group("/api")

	histogramController := NewHistogramController(reportDB)

	deviceController := NewDeviceController(mainDB)

	credentialProfileController := NewCredentialProfileController(mainDB)

	discoveryProfileController := NewDiscoveryProfileController(mainDB)

	// Use POST for histogram queries with body parameters
	api.POST("/histogram", histogramController.GetHistogram)

	// Keep the GET endpoint temporarily for backward compatibility
	// but mark it as deprecated in API docs
	api.GET("/histogram", func(c *gin.Context) {

		c.JSON(400, gin.H{

			"error": "GET /api/histogram is deprecated. Please use POST /api/histogram with JSON body",

			"example": HistogramQueryRequest{

				From: 1744610677,

				To: 1744620677,

				CounterID: 1,

				ObjectIDs: []uint32{169093227, 2130706433},
			},
		})

	})

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
