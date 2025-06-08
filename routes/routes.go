// routes/user_routes.go
package routes

import (
	"v01_system_backend/apps/controllers"
	"v01_system_backend/apps/middleware"

	"github.com/labstack/echo/v4"
)

// SetupUserRoutes mengatur semua endpoint untuk user CRUD
// ENDPOINT MAPPING:
// GET    /users          -> GetAll (list users dengan pagination)
// GET    /users/:id      -> GetByID (detail user)
// POST   /users          -> Create (buat user baru)
// PUT    /users/:id      -> Update (update user)
// DELETE /users/:id      -> Delete (soft delete user)
// PUT    /users/:id/status -> UpdateStatus (ubah status user)
// PUT    /users/:id/reset-password -> ResetPassword (reset password)
//
// ROLE MANAGEMENT:
// GET    /users/:id/roles -> GetUserRoles (ambil role user)
// POST   /users/:id/roles -> AssignRoles (assign role ke user)
// DELETE /users/:id/roles -> RemoveRoles (remove role dari user)
//
// PERMISSIONS & ACTIVITIES:
// GET    /users/:id/permissions -> GetUserPermissions (ambil permissions)
// GET    /users/:id/activities  -> GetUserActivities (ambil activity log)
func SetupUserRoutes(g *echo.Group, userController *controllers.UserController, authMiddleware *middleware.AuthMiddleware) {
	// Group semua endpoint user di bawah /users
	users := g.Group("/users")

	// === BASIC CRUD OPERATIONS ===
	// Setiap endpoint menggunakan permission middleware untuk authorization
	users.GET("", userController.GetAll, authMiddleware.RequirePermission("users.read"))
	users.GET("/:id", userController.GetByID, authMiddleware.RequirePermission("users.read"))
	users.POST("", userController.Create, authMiddleware.RequirePermission("users.create"))
	users.PUT("/:id", userController.Update, authMiddleware.RequirePermission("users.update"))
	users.DELETE("/:id", userController.Delete, authMiddleware.RequirePermission("users.delete"))

	// === USER STATUS & PASSWORD MANAGEMENT ===
	users.PUT("/:id/status", userController.UpdateStatus, authMiddleware.RequirePermission("users.update"))
	users.PUT("/:id/reset-password", userController.ResetPassword, authMiddleware.RequirePermission("users.reset_password"))

	// === ROLE MANAGEMENT ===
	// Group endpoint untuk role management di bawah /users/:id/roles
	roles := users.Group("/:id/roles")
	roles.GET("", userController.GetUserRoles, authMiddleware.RequirePermission("users.read"))
	roles.POST("", userController.AssignRoles, authMiddleware.RequirePermission("users.manage"))
}
