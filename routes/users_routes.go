package routes

import (
	"database/sql"
	"time"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupUserRoutes(api *echo.Group, db *sql.DB) {
	userController := controller.NewUserController(db)

	// Add request logging middleware
	api.Use(middleware.Logger())

	// Add request timeout middleware
	api.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))

	// User CRUD routes
	users := api.Group("/users")
	users.POST("", userController.CreateUser)       // Create user
	users.GET("", userController.GetAllUsers)       // Get all users with pagination & filtering
	users.GET("/:id", userController.GetUser)       // Get user by ID
	users.PUT("/:id", userController.UpdateUser)    // Update user
	users.DELETE("/:id", userController.DeleteUser) // Delete user (soft delete)

	// Additional user routes
	users.GET("/status/:status_id", userController.GetUsersByStatus) // Get users by status
	users.GET("/search", userController.SearchUsers)                 // Search users

	// // Utility routes
	// users.GET("/check-username", userController.CheckUsernameAvailability) // Check username availability
	// users.GET("/check-email", userController.CheckEmailAvailability)       // Check email availability

	// // Bulk operations
	// users.PUT("/bulk-update", userController.BulkUpdateUsers)    // Bulk update users
	// users.DELETE("/bulk-delete", userController.BulkDeleteUsers) // Bulk delete users

	// // Password management
	// users.PUT("/:id/change-password", userController.ChangePassword) // Change password
	// users.POST("/:id/reset-password", userController.ResetPassword)  // Reset password

	// // Export routes
	// export := users.Group("/export")
	// export.GET("/csv", userController.ExportUsersCSV)     // Export to CSV
	// export.GET("/excel", userController.ExportUsersExcel) // Export to Excel
	// export.GET("/pdf", userController.ExportUsersPDF)     // Export to PDF

	// // Auth routes
	// auth := api.Group("/auth")
	// auth.POST("/login", userController.Login)          // User login
	// auth.POST("/logout", userController.Logout)        // User logout
	// auth.POST("/refresh", userController.RefreshToken) // Refresh JWT token
}
