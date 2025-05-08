package main

import (
	"context"
	"go.uber.org/zap"
	"net/http"
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

	shutdownChannel := InitShutdownHandler(1)

	// Initialize the reportDb client
	reportDB, err := InitReportDBClient()

	if err != nil {

		Logger.Error("error initializing reportDB:", zap.Error(err))

		return
	}

	// Initialize configDb client
	configDB, err := InitConfigDBClient(GetConfigDBConnectionString())

	if err != nil {

		Logger.Error("error initializing configDB:", zap.Error(err))

		return

	}

	// Initialize the provisioning publisher
	provisioningPublisher, err := InitProvisioningPublisher()

	if err != nil {

		Logger.Error("error initializing provisioningPublisher:", zap.Error(err))

		return

	}

	// polled data router
	pollDataRouter := InitPollDataRouter()

	// Initialize router & server
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

	server := &http.Server{

		Addr: ":" + ServerPort,

		Handler: router,
	}

	// Start the server in a separate routine

	go func() {

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {

			Logger.Error("error starting server:", zap.Error(err))

		}

	}()

	// Wait for shutdown signal

	<-shutdownChannel

	pollDataRouter.Close()

	provisioningPublisher.Close()

	reportDB.Close()

	configDB.Close()

	// Close server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {

		Logger.Error("error shutting down server:", zap.Error(err))

	}

	Logger.Info("Server stopped")

}
