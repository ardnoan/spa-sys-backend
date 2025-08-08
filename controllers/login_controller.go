package controller

import (
	"net/http"
	"strings"
	"v01_system_backend/services"

	"github.com/labstack/echo/v4"
)

type LoginController struct {
	authService *services.AuthService
}

func NewLoginController(authService *services.AuthService) *LoginController {
	return &LoginController{
		authService: authService,
	}
}

func (lc *LoginController) Login(c echo.Context) error {
	var req services.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
		})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Validation failed",
		})
	}

	// Get client info
	clientIP := c.RealIP()
	userAgent := c.Request().UserAgent()

	// Call service (which calls stored procedure)
	response, err := lc.authService.Login(req, clientIP, userAgent)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	if !response.Success {
		return c.JSON(http.StatusUnauthorized, response)
	}

	return c.JSON(http.StatusOK, response)
}

func (lc *LoginController) Logout(c echo.Context) error {
	tokenString := lc.extractTokenFromHeader(c)
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authorization header required",
		})
	}

	clientIP := c.RealIP()
	userAgent := c.Request().UserAgent()

	// Call service (which calls stored procedure)
	err := lc.authService.Logout(tokenString, clientIP, userAgent)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	})
}

func (lc *LoginController) GetCurrentUser(c echo.Context) error {
	// Get user info from context (set by middleware)
	userID := c.Get("user_id")
	username := c.Get("username")

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":  userID,
		"username": username,
	})
}

func (lc *LoginController) RefreshToken(c echo.Context) error {
	tokenString := lc.extractTokenFromHeader(c)
	if tokenString == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authorization header required",
		})
	}

	// For now, we'll implement simple token refresh
	// In production, you might want a separate refresh token mechanism
	claims, err := lc.authService.ValidateToken(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid token",
		})
	}

	// Generate new token (this is simplified - in production use refresh tokens)
	// This would need to be implemented in the service
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Token refresh not yet implemented",
		"user_id": claims.UserID,
	})
}

func (lc *LoginController) extractTokenFromHeader(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(authHeader, "Bearer ")
}
