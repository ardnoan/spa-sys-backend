package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupUserRoutes(api *echo.Group, db *sql.DB) {
	userController := controller.NewUserController(db)

	// User routes
	users := api.Group("/users")
	users.POST("", userController.CreateUser)
	users.GET("", userController.GetAllUsers)
	users.GET("/:id", userController.GetUser)
	users.PUT("/:id", userController.UpdateUser)
	users.DELETE("/:id", userController.DeleteUser)

	// Auth routes
	auth := api.Group("/auth")
	auth.POST("/login", userController.Login)
}
