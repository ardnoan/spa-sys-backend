package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupUsersActivityLogsRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewUsersActivityLogsController(db)
	Routes := api.Group("/users-activity-logs")
	Routes.GET("", Controllers.GetAllUsersActivityLogs)
}
