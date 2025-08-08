package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupMenusRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewMenusController(db)
	routes := api.Group("/menus")

	// User-specific menu routes using procedures
	routes.GET("/user", Controllers.GetUserMenus)                               // GET /api/menus/user?user_id=1
	routes.GET("/user/root", Controllers.GetRootMenusForUser)                   // GET /api/menus/user/root?user_id=1
	routes.GET("/:id/breadcrumb", Controllers.GetMenuBreadcrumb)                // GET /api/menus/1/breadcrumb?user_id=1
	routes.GET("/:parent_id/children", Controllers.GetChildMenusForUser)        // GET /api/menus/1/children?user_id=1
	routes.GET("/:parent_id/descendants", Controllers.GetAllDescendantsForUser) // GET /api/menus/1/descendants?user_id=1
}
