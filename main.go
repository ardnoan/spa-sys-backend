package main

import (
	"log"
	"v01_system_backend/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize config
	InitConfig()

	// Initialize database
	InitDatabase()
	defer DB.Close()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Setup routes with DB
	routes.SetupRoutes(e, DB)

	// Start server
	log.Printf("Server starting on port %s", AppConfig.ServerPort)
	e.Logger.Fatal(e.Start(":" + AppConfig.ServerPort))
}
