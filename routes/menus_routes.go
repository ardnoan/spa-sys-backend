package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupMenusRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewMenusController(db)
	routes := api.Group("/menus")
	routes.GET("/by-route/:route", Controllers.GetMenuByRoute)

	// Basic CRUD routes
	routes.GET("", Controllers.GetAllMenus)                       // GET /api/menus
	routes.GET("/root", Controllers.GetRootMenus)                 // GET /api/menus/root
	routes.GET("/:id", Controllers.GetMenuById)                   // GET /api/menus/1
	routes.GET("/:id/breadcrumb", Controllers.GetMenuBreadcrumb)  // GET /api/menus/1/breadcrumb
	routes.GET("/:parent_id/children", Controllers.GetChildMenus) // GET /api/menus/1/children
	routes.GET("/hierarchical", Controllers.GetHierarchicalMenus) // GET /api/menus/hierarchical

	// Basic CRUD routes
	routes.POST("", Controllers.CreateMenu)       // POST /api/menus
	routes.PUT("/:id", Controllers.UpdateMenu)    // PUT /api/menus/1
	routes.DELETE("/:id", Controllers.DeleteMenu) // DELETE /api/menus/1

	routes.GET("/:parent_id/descendants", Controllers.GetAllDescendants)
}
