package middleware

import (
	"net/http"
	"strings"
	"v01_system_backend/services"

	"github.com/labstack/echo/v4"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (am *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get token from Authorization header
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Authorization header required",
			})
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate JWT token
		claims, err := am.authService.ValidateToken(tokenString)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Invalid token",
			})
		}

		// Validate session in database (calls stored procedure)
		sessionValidation, err := am.authService.ValidateSession(tokenString)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Session validation error",
			})
		}

		if !sessionValidation.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": sessionValidation.Message,
			})
		}

		// Add user info to context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)

		return next(c)
	}
}
