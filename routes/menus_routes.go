package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupMenusRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewMenusController(db)
	routes := api.Group("/menus")
	routes.GET("", Controllers.GetAllMenus)
	routes.GET("/root", Controllers.GetRootMenus)
}
