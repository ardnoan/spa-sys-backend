package routes

import (
	"database/sql"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, db *sql.DB) {
	// API version 1
	api := e.Group("/api/v1")

	// Setup user routes
	SetupUserRoutes(api, db)
	SetupDepartmentRoutes(api, db)
	SetupStatusRoutes(api, db)
	SetupRoleRoutes(api, db)
	SetupPermissionRoutes(api, db)
	SetupEmailTemplatesRoutes(api, db)
	SetupMenusRoutes(api, db)
	SetupNotficationRoutes(api, db)
	SetupPasswordResetTokensRoutes(api, db)
	SetupRolesMenusRoutes(api, db)
	SetupRolesPermissionsRoutes(api, db)
	SetupSystemsSettingsRoutes(api, db)
	SetupUsersSessionsRoutes(api, db)
	SetupUsersActivityLogsRoutes(api, db)
	SetupUsersPasswordHistoryRoutes(api, db)
	SetupUsersRolesRoutes(api, db)
	SetupAuthRoutes(api, db)

	// Health check
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status":  "OK",
			"message": "Server is running",
		})
	})
}
