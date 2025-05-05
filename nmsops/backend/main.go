package main

import (
	"go.uber.org/zap"
	. "nms-backend/db"
	. "nms-backend/router"
	. "nms-backend/services"
	. "nms-backend/utils"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize configuration
	err := LoadConfig()

	if err != nil {

		Logger.Error("error loading config:", zap.Error(err))

		return

	}

	// Initialize the reportDb client
	reportDB, err := InitReportDBClient()

	if err != nil {

		Logger.Error("error initializing reportDB:", zap.Error(err))

		return
	}

	defer reportDB.Shutdown()

	// Initialize configDb client
	configDB, err := InitConfigDBClient(GetConfigDBConnectionString())

	if err != nil {

		Logger.Error("error initializing configDB:", zap.Error(err))

		return

	}

	defer configDB.Close()

	// Initialize the provisioning publisher
	provisioningPublisher, err := InitProvisioningPublisher()

	if err != nil {

		Logger.Error("error initializing provisioningPublisher:", zap.Error(err))

		return

	}

	defer provisioningPublisher.Close()

	// polled data router

	context := InitPollRouter()

	defer context.Term()

	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{

		AllowOrigins: []string{"*"},

		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},

		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},

		ExposeHeaders: []string{"Content-Length"},

		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	}))

	SetupRoutes(router, reportDB, configDB, provisioningPublisher)

	Logger.Info("Server started at port 8080")

	if err := router.Run(":8080"); err != nil {

		Logger.Error("Server exited with error:", zap.Error(err))

	}
}
