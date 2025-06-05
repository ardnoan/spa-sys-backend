package routes

import (
	"v01_system_backend/apps/controllers"

	"github.com/labstack/echo/v4"
)

func SetupSystemRoutes(g *echo.Group, systemController *controllers.SystemController) {
	system := g.Group("/system")

	// System settings
	settings := system.Group("/settings")
	settings.GET("", systemController.GetAllSettings)
	settings.GET("/public", systemController.GetPublicSettings)
	settings.GET("/:key", systemController.GetSettingByKey)
	settings.POST("", systemController.CreateSetting)
	settings.PUT("/:key", systemController.UpdateSetting)
	settings.DELETE("/:key", systemController.DeleteSetting)

	// Application info
	system.GET("/info", systemController.GetAppInfo)
	system.GET("/health", systemController.HealthCheck)
}

func SetupMenuRoutes(g *echo.Group, menuController *controllers.MenuController) {
	menus := g.Group("/menus")

	menus.GET("", menuController.GetAll)
	menus.GET("/tree", menuController.GetMenuTree)
	menus.GET("/user-menus", menuController.GetUserMenus)
	menus.GET("/:id", menuController.GetByID)
	menus.POST("", menuController.Create)
	menus.PUT("/:id", menuController.Update)
	menus.DELETE("/:id", menuController.Delete)
	menus.PUT("/:id/order", menuController.UpdateOrder)
}
