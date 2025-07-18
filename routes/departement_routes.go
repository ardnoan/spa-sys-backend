package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"

	"v01_system_backend/repositories"
	"v01_system_backend/services"
)

func SetupDepartmentRoutes(api *echo.Group, db *sql.DB) {
	// Initialize layers
	departmentRepo := repositories.NewDepartmentRepository(db)
	departmentService := services.NewDepartmentService(departmentRepo)
	departmentController := controller.NewDepartmentController(departmentService)

	// Department CRUD routes
	departments := api.Group("/departments")
	departments.POST("", departmentController.CreateDepartment)
	departments.GET("", departmentController.GetAllDepartments)
	departments.GET("/hierarchy", departmentController.GetDepartmentHierarchy)
	departments.GET("/search", departmentController.SearchDepartments)
	departments.GET("/:id", departmentController.GetDepartment)
	departments.PUT("/:id", departmentController.UpdateDepartment)
	departments.DELETE("/:id", departmentController.DeleteDepartment)
	departments.GET("/:id/users", departmentController.GetUsersByDepartment)
}
