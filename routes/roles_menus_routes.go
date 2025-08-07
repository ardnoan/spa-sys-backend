package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupRolesMenusRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewRoleMenusController(db)
	Routes := api.Group("/roles-menus")

	// CRUD routes
	Routes.GET("", Controllers.GetAllRoleMenus)       // GET /api/roles-menus
	Routes.GET("/:id", Controllers.GetRoleMenuByID)   // GET /api/roles-menus/:id
	Routes.POST("", Controllers.CreateRoleMenu)       // POST /api/roles-menus
	Routes.PUT("/:id", Controllers.UpdateRoleMenu)    // PUT /api/roles-menus/:id
	Routes.DELETE("/:id", Controllers.DeleteRoleMenu) // DELETE /api/roles-menus/:id

	// Additional utility routes
	Routes.POST("/bulk-update", Controllers.BulkUpdatePermissions) // POST /api/roles-menus/bulk-update
	Routes.POST("/copy-permissions", Controllers.CopyPermissions)  // POST /api/roles-menus/copy-permissions

	// Helper routes for dropdowns (can also be separate endpoints)
	Routes.GET("/users-roles", Controllers.GetAllRoles) // GET /api/users-roles
	Routes.GET("/menus", Controllers.GetAllMenus)       // GET /api/menus
}
