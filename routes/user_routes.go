package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupUserRoutes(api *echo.Group, db *sql.DB) {
	userController := controller.NewUserController(db)

	// User CRUD routes
	users := api.Group("/users")
	users.POST("", userController.CreateUser)       // Create user
	users.GET("", userController.GetAllUsers)       // Get all users with pagination
	users.GET("/:id", userController.GetUser)       // Get user by ID
	users.PUT("/:id", userController.UpdateUser)    // Update user
	users.DELETE("/:id", userController.DeleteUser) // Delete user (soft delete)

	// Additional user routes
	users.GET("/status/:status_id", userController.GetUsersByStatus) // Get users by status
	users.GET("/search", userController.SearchUsers)                 // Search users

	// Auth routes
	auth := api.Group("/auth")
	auth.POST("/login", userController.Login)
}
