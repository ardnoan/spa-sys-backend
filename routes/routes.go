package routes

import (
	"v01_system_backend/apps/handlers"
	"v01_system_backend/apps/middleware"
	"v01_system_backend/apps/repositories"
	"v01_system_backend/apps/services"
	"v01_system_backend/config"

	"github.com/labstack/echo/v4"
)

func SetupRoutes(e *echo.Echo, cfg *config.Config) {
	// Get database instance
	db := config.GetDB()

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	sessionRepo := repositories.NewSessionRepository(db)
	activityRepo := repositories.NewActivityRepository(db)

	// Initialize services
	authService := services.NewAuthService(userRepo, sessionRepo, cfg)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)
	loggingMiddleware := middleware.NewLoggingMiddleware(activityRepo)

	// Apply global middleware
	e.Use(middleware.CORS())
	e.Use(middleware.RequestLogger())

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "OK", "service": "auth"})
	})

	// API version 1
	v1 := e.Group("/api/v1")

	// Setup auth routes (includes both public and protected)
	SetupAuthRoutes(v1, authHandler)

	// Apply auth middleware to protected routes
	protected := v1.Group("", authMiddleware.Authenticate)
	protected.Use(loggingMiddleware.ActivityLogger())

	// Add other protected routes here
	// SetupUserRoutes(protected, userHandler)
}
