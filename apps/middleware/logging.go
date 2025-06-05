package middleware

import (
	"time"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/repositories"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type LoggingMiddleware struct {
	activityRepo *repositories.ActivityRepository
}

func NewLoggingMiddleware(activityRepo *repositories.ActivityRepository) *LoggingMiddleware {
	return &LoggingMiddleware{
		activityRepo: activityRepo,
	}
}

func RequestLogger() echo.MiddlewareFunc {
	return middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format:           "${time_rfc3339} ${method} ${uri} ${status} ${latency_human} ${bytes_in}/${bytes_out}\n",
		CustomTimeFormat: "2006-01-02 15:04:05",
	})
}

func (m *LoggingMiddleware) ActivityLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Call next handler
			err := next(c)

			// Log activity if user is authenticated
			if userID := c.Get("user_id"); userID != nil {
				go m.logActivity(c, start, userID.(int))
			}

			return err
		}
	}
}

func (m *LoggingMiddleware) logActivity(c echo.Context, start time.Time, userID int) {
	// Skip logging for certain paths
	skipPaths := []string{"/health", "/metrics"}
	for _, path := range skipPaths {
		if c.Request().URL.Path == path {
			return
		}
	}

	activity := &models.UserActivityLog{
		UserID:         &userID,
		Action:         c.Request().Method,
		TargetType:     ptrString("HTTP_REQUEST"),
		MenuName:       ptrString(c.Request().URL.Path),
		Description:    ptrString(c.Request().Method + " " + c.Request().URL.Path),
		IPAddress:      ptrString(c.RealIP()),
		UserAgent:      ptrString(c.Request().UserAgent()),
		ResponseStatus: ptrInt(c.Response().Status),
	}

	// This would be called asynchronously to avoid blocking the request
	m.activityRepo.Create(activity)
}

func ptrString(s string) *string {
	return &s
}

func ptrInt(i int) *int {
	return &i
}
