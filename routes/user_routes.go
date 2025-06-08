// routes/user_routes.go
package routes

import (
	"v01_system_backend/apps/controllers"
	"v01_system_backend/apps/middleware"

	"github.com/labstack/echo/v4"
)

func SetupUserRoutes(g *echo.Group, userController *controllers.UserController, authMiddleware *middleware.AuthMiddleware) {
	users := g.Group("/users")

	// Apply permission middleware for different operations
	users.GET("", userController.GetAll, authMiddleware.RequirePermission("users.read"))
	users.GET("/:id", userController.GetByID, authMiddleware.RequirePermission("users.read"))
	users.POST("", userController.Create, authMiddleware.RequirePermission("users.create"))
	users.PUT("/:id", userController.Update, authMiddleware.RequirePermission("users.update"))
	users.DELETE("/:id", userController.Delete, authMiddleware.RequirePermission("users.delete"))
	users.PUT("/:id/status", userController.UpdateStatus, authMiddleware.RequirePermission("users.update"))
	users.PUT("/:id/reset-password", userController.ResetPassword, authMiddleware.RequirePermission("users.reset_password"))

	// User roles management
	roles := users.Group("/:id/roles")
	roles.GET("", userController.GetUserRoles, authMiddleware.RequirePermission("users.read"))
	roles.POST("", userController.AssignRoles, authMiddleware.RequirePermission("users.manage_roles"))
	roles.DELETE("", userController.RemoveRoles, authMiddleware.RequirePermission("users.manage_roles"))

	// User permissions
	users.GET("/:id/permissions", userController.GetUserPermissions, authMiddleware.RequirePermission("users.read"))

	// User activities
	users.GET("/:id/activities", userController.GetUserActivities, authMiddleware.RequirePermission("users.read"))
}
