package routes

import (
	"v01_system_backend/apps/controllers"

	"github.com/labstack/echo/v4"
)

func SetupAuthRoutes(g *echo.Group, authController *controllers.AuthController) {
	auth := g.Group("/auth")

	auth.POST("/login", authController.Login)
	auth.POST("/logout", authController.Logout)
	auth.POST("/refresh", authController.RefreshToken)
	auth.POST("/change-password", authController.ChangePassword)
	auth.GET("/profile", authController.GetProfile)
}
