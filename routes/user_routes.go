// routes/user_routes.go
package routes

import (
	"v01_system_backend/apps/controllers"

	"github.com/labstack/echo/v4"
)

func SetupUserRoutes(g *echo.Group, userController *controllers.UserController) {
	users := g.Group("/users")

	users.GET("", userController.GetAll)
	users.GET("/:id", userController.GetByID)
	users.POST("", userController.Create)
	users.PUT("/:id", userController.Update)
	users.DELETE("/:id", userController.Delete)
	users.PUT("/:id/status", userController.UpdateStatus)

	// User roles
	users.GET("/:id/roles", userController.GetUserRoles)
	users.POST("/:id/roles", userController.AssignRoles)
	users.DELETE("/:id/roles", userController.RemoveRoles)

	// User permissions
	users.GET("/:id/permissions", userController.GetUserPermissions)

	// User activity logs
	users.GET("/:id/activities", userController.GetUserActivities)
}
