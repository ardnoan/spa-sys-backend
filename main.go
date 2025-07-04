package main

import (
	"log"
	"net/http"
	"v01_system_backend/config"
	"v01_system_backend/routes"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// CustomValidator wraps the validator
type CustomValidator struct {
	validator *validator.Validate
}

// Validate validates structs
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	// Initialize config
	config.InitConfig()

	// Initialize database
	config.InitDatabase()
	defer config.DB.Close()

	// Create Echo instance
	e := echo.New()

	// Set custom validator
	e.Validator = &CustomValidator{validator: validator.New()}

	// Middleware
	// e.Use(middleware.Logger())
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
