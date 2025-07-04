package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupPasswordResetTokensRoutes(api *echo.Group, db *sql.DB) {
	Controllers := controller.NewPasswordResetTokensController(db)
	Routes := api.Group("/password-reset-tokens")
	Routes.GET("", Controllers.GetAllPasswordResetTokens)
}
