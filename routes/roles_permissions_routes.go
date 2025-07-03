package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupRolesPermissionsRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewRolePermissionsController(db)
	Routes := api.Group("/email-templates")
	Routes.GET("", Controllers.GetAllRolePermissions)
}
