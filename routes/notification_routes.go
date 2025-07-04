package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupNotficationRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewNotificationsController(db)
	Routes := api.Group("/notifications")
	Routes.GET("", Controllers.GetAllNotifications)
}
