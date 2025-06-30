package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupPermissionRoutes(api *echo.Group, db *sql.DB) {
	permissionController := controller.NewPermissionController(db)

	permissions := api.Group("/permissions")

	// Permission CRUD routes
	permissions.POST("", permissionController.CreatePermission)       // Create permission
	permissions.GET("", permissionController.GetAllPermissions)       // Get all permissions with pagination and filters
	permissions.GET("/:id", permissionController.GetPermission)       // Get permission by ID
	permissions.PUT("/:id", permissionController.UpdatePermission)    // Update permission
	permissions.DELETE("/:id", permissionController.DeletePermission) // Delete permission (soft delete)

	// Additional permission routes
	permissions.GET("/options", permissionController.GetPermissionOptions)     // Get permission options for dropdowns
	permissions.GET("/modules", permissionController.GetModules)               // Get available modules
	permissions.GET("/check-code", permissionController.CheckCodeAvailability) // Check if permission code is available
	permissions.GET("/search", permissionController.SearchPermissions)         // Search permissions
}
