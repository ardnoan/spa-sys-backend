// routes/routes.go
package routes

import (
	"v01_system_backend/apps/controllers"
	"v01_system_backend/apps/middleware"
	"v01_system_backend/apps/repositories"
	"v01_system_backend/apps/services"
	"v01_system_backend/config"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, cfg *config.Config) {
	// Get database instance
	db := config.GetDB()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	menuRepo := repositories.NewMenuRepository(db)
	systemRepo := repositories.NewSystemRepository(db)
	activityRepo := repositories.NewActivityRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg)
	userService := services.NewUserService(userRepo)
	roleService := services.NewRoleService(roleRepo)
	menuService := services.NewMenuService(menuRepo)
	systemService := services.NewSystemService(systemRepo)

	// Initialize controllers
	authController := controllers.NewAuthController(authService)
	userController := controllers.NewUserController(userService)
	roleController := controllers.NewRoleController(roleService)
	menuController := controllers.NewMenuController(menuService)
	systemController := controllers.NewSystemController(systemService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	loggingMiddleware := middleware.NewLoggingMiddleware(activityRepo)

	// Apply global middleware
	e.Use(middleware.CORS())
	e.Use(middleware.RequestLogger())
	e.Use(loggingMiddleware.ActivityLogger())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "OK"})
	})

	// API version 1
	v1 := e.Group("/api/v1")

	// Public routes (no authentication required)
	public := v1.Group("")
	SetupAuthRoutes(public, authController)

	// Protected routes (authentication required)
	protected := v1.Group("", authMiddleware.Authenticate)
	SetupUserRoutes(protected, userController)
	SetupRoleRoutes(protected, roleController)
	SetupMenuRoutes(protected, menuController)
	SetupSystemRoutes(protected, systemController)
}
