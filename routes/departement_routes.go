package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupDepartmentRoutes(api *echo.Group, db *sql.DB) {
	departmentController := controller.NewDepartmentController(db)

	// Department CRUD routes
	departments := api.Group("/departments")
	departments.POST("", departmentController.CreateDepartment)       // Create department
	departments.GET("", departmentController.GetAllDepartments)       // Get all departments with pagination and filters
	departments.GET("/:id", departmentController.GetDepartment)       // Get department by ID
	departments.PUT("/:id", departmentController.UpdateDepartment)    // Update department
	departments.DELETE("/:id", departmentController.DeleteDepartment) // Delete department (soft delete)

	// Additional department routes
	departments.GET("/hierarchy", departmentController.GetDepartmentHierarchy) // Get department hierarchy
	departments.GET("/:id/users", departmentController.GetUsersByDepartment)   // Get users in department
	departments.GET("/search", departmentController.SearchDepartments)         // Search departments
}
