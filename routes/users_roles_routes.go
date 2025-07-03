package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupUsersRolesRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewUserRoleController(db)
	Routes := api.Group("/email-templates")
	Routes.GET("", Controllers.GetAllUserRoles)
}
