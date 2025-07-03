package routes

import (
	"database/sql"
	controller "v01_system_backend/controllers"

	"github.com/labstack/echo/v4"
)

func SetupEmailTemplatesRoutes(api *echo.Group, db *sql.DB) {
	emailController := controller.NewEmailTemplatesController(db)
	emailTemplates := api.Group("/email-templates")
	emailTemplates.GET("", emailController.GetAllEmailTemplates)
}
