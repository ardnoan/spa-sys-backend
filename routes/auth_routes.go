package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"
	"v01_system_backend/services"

	"github.com/labstack/echo/v4"
)

func SetupAuthRoutes(api *echo.Group, db *sql.DB) {
	// Initialize services
	authService := services.NewAuthService(db)

	// Initialize controllers
	loginController := controller.NewLoginController(authService)

	auth := api.Group("/auth")
	auth.POST("/login", loginController.Login)
	auth.POST("/logout", loginController.Logout)
	auth.GET("/me", loginController.GetCurrentUser)
	auth.POST("/refresh", loginController.RefreshToken)
}
