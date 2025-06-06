package main

import (
	"log"
	"v01_system_backend/config"
	"v01_system_backend/routes"

	"github.com/labstack/echo/v4"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	config.InitDB(cfg)

	// Create Echo instance
	e := echo.New()

	// Setup routes
	routes.SetupRoutes(e, cfg)

	// Start server
	port := ":" + cfg.Server.Port
	log.Printf("Server starting on port %s", port)
	log.Fatal(e.Start(port))
}
