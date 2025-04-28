package main

import (
	"log"
	"nms-backend/config"
	"nms-backend/db"
	"nms-backend/routes"
	"nms-backend/services"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize configuration
	cfg := config.NewConfig()

	// Initialize the reportDb client

	reportDB, err := db.InitReportDbClient()

	if err != nil {

		log.Fatal("Failed to initialize report DB", err)

	}

	defer reportDB.Shutdown()

	// Initialize configDb client

	configDB, err := db.NewConfigDB(cfg.GetDBConnectionString())

	if err != nil {

		log.Fatal("Failed to initialize main DB", err)
		
	}

	defer configDB.Close()

	// Initialize the provisioning publisher

	provisioningPublisher, err := services.InitProvisioningPublisher()

	if err != nil {

		log.Fatal("Failed to initialize provisioning publisher", err)

	}

	defer provisioningPublisher.Close()

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

	routes.SetupRoutes(router, reportDB, configDB, provisioningPublisher)

	log.Println("Server starting at :8080")

	if err := router.Run(":8080"); err != nil {

		log.Fatal("Server exited with error: ", err)

	}
}
