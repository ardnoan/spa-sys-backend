package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupUsersSessionsRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewUserSessionsController(db)
	Routes := api.Group("/users-sessions")
	Routes.GET("", Controllers.GetAllUserSessions)
}
