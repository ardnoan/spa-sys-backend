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
	// Initialize repositories
	userRepo := repositories.NewUserRepository(config.DB)
	roleRepo := repositories.NewRoleRepository(config.DB)
	menuRepo := repositories.NewMenuRepository(config.DB)
	systemRepo := repositories.NewSystemRepository(config.DB)

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
