package main

import (
	"log"
	"nms-backend/config"
	"nms-backend/db"
	"nms-backend/reportdb"
	"nms-backend/routes"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize configuration
	cfg := config.NewConfig()

	reportDB, err := reportdb.InitClient()

	if err != nil {
		log.Fatalf("Failed to initialize report DB: %v", err)
	}

	defer reportDB.Shutdown()

	configDB, err := db.NewConfigDB(cfg.GetDBConnectionString())

	if err != nil {
		log.Fatalf("Failed to initialize main DB: %v", err)
	}

	defer configDB.Close()

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

	routes.SetupRoutes(router, reportDB, configDB)

	log.Println("Server starting at :8080")

	if err := router.Run(":8080"); err != nil {

		log.Fatal("Server exited with error: ", err)

	}
}
