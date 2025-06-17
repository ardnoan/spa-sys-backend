package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupStatusRoutes(api *echo.Group, db *sql.DB) {
	statusController := controller.NewStatusController(db)

	// Status CRUD routes
	statuses := api.Group("/statuses")
	statuses.POST("", statusController.CreateStatus)       // Create status
	statuses.GET("", statusController.GetAllStatuses)      // Get all statuses with pagination and filters
	statuses.GET("/:id", statusController.GetStatus)       // Get status by ID
	statuses.PUT("/:id", statusController.UpdateStatus)    // Update status
	statuses.DELETE("/:id", statusController.DeleteStatus) // Delete status (soft delete)

	// Additional status routes
	statuses.GET("/search", statusController.SearchStatuses) // Search statuses
}
