package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupRolesMenusRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewRoleMenusController(db)
	Routes := api.Group("/roles-menus")
	Routes.GET("", Controllers.GetAllRoleMenus)
}
