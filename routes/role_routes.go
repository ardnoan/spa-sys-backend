package routes

import (
	"v01_system_backend/apps/controllers"

	"github.com/labstack/echo/v4"
)

func SetupRoleRoutes(g *echo.Group, roleController *controllers.RoleController) {
	roles := g.Group("/roles")

	roles.GET("", roleController.GetAll)
	roles.GET("/:id", roleController.GetByID)
	roles.POST("", roleController.Create)
	roles.PUT("/:id", roleController.Update)
	roles.DELETE("/:id", roleController.Delete)

	// Role permissions
	roles.GET("/:id/permissions", roleController.GetRolePermissions)
	roles.POST("/:id/permissions", roleController.AssignPermissions)
	roles.DELETE("/:id/permissions", roleController.RemovePermissions)

	// Role menus
	roles.GET("/:id/menus", roleController.GetRoleMenus)
	roles.POST("/:id/menus", roleController.AssignMenus)
}
