package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupSystemsSettingsRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewSystemSettingsController(db)
	Routes := api.Group("/systems-settings")
	Routes.GET("", Controllers.GetAllSystemSettings)
}
