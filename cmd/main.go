// main.go
package main

import (
	"log"
	"v01_system_backend/config"
	"v01_system_backend/routes"

	"github.com/labstack/echo/v4"
)

func main() {
	// Load configuration
	cfg := config.Load() // Now this matches

	// Initialize database
	config.InitDB(cfg)

	// Auto-migrate your models here if needed
	// config.DB.AutoMigrate(&models.User{}, &models.Role{}, &models.UserActivityLog{})

	// Create Echo instance
	e := echo.New()

	// Setup routes
	routes.SetupRoutes(e, cfg)

	// Start server
	log.Fatal(e.Start(":" + cfg.Server.Port))
}
