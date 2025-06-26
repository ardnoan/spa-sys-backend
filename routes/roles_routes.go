package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupRoleRoutes(api *echo.Group, db *sql.DB) {
	roleController := controller.NewRoleController(db)

	roles := api.Group("/roles")

	// Role CRUD routes
	roles.POST("", roleController.CreateRole)       // Create role
	roles.GET("", roleController.GetAllRoles)       // Get all roles with pagination and filters
	roles.GET("/:id", roleController.GetRole)       // Get role by ID
	roles.PUT("/:id", roleController.UpdateRole)    // Update role
	roles.DELETE("/:id", roleController.DeleteRole) // Delete role (soft delete)

	// Additional role routes
	roles.GET("/options", roleController.GetRoleOptions)           // Get role options
	roles.GET("/check-code", roleController.CheckCodeAvailability) // Check if role code is available
}
