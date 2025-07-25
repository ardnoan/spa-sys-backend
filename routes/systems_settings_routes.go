package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupSystemsSettingsRoutes(api *echo.Group, db *sql.DB) {
	controller := controller.NewSystemSettingsController(db)
	routes := api.Group("/systems-settings")

	// CRUD Routes
	routes.GET("", controller.GetAllSystemSettings)       // GET /api/systems-settings
	routes.GET("/:id", controller.GetSystemSettingByID)   // GET /api/systems-settings/:id
	routes.POST("", controller.CreateSystemSetting)       // POST /api/systems-settings
	routes.PUT("/:id", controller.UpdateSystemSetting)    // PUT /api/systems-settings/:id
	routes.DELETE("/:id", controller.DeleteSystemSetting) // DELETE /api/systems-settings/:id

	// Public settings route (for frontend consumption)
	routes.GET("/public", controller.GetPublicSettings) // GET /api/systems-settings/public
}
