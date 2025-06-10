package main

import (
	"log"
	"v01_system_backend/config"
	"v01_system_backend/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize config
	config.InitConfig()

	// Initialize database
	config.InitDatabase()
	defer config.DB.Close()

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Setup routes with DB
	routes.SetupRoutes(e, config.DB)

	// Optionally: basic ping test
	e.GET("/ping", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "pong",
		})
	})

	// Start server
	log.Printf("Server starting on port %s", config.AppConfig.ServerPort)
	e.Logger.Fatal(e.Start(":" + config.AppConfig.ServerPort))
}
