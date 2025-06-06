package routes

import (
	"v01_system_backend/apps/handlers"

	"github.com/labstack/echo/v4"
)

func SetupAuthRoutes(g *echo.Group, authHandler *handlers.AuthHandler) {
	auth := g.Group("/auth")

	// Public routes (no authentication required)
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.RefreshToken)
	auth.POST("/forgot-password", authHandler.ForgotPassword)
	auth.POST("/reset-password", authHandler.ResetPassword)

	// Protected routes (authentication required)
	auth.POST("/logout", authHandler.Logout)
	auth.POST("/change-password", authHandler.ChangePassword)
	auth.GET("/profile", authHandler.GetProfile)
	auth.GET("/me", authHandler.GetMe)
	auth.PUT("/profile", authHandler.UpdateProfile)
}
