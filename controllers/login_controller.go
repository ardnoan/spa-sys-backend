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
		return err
	}

	// Get client IP and User Agent
	clientIP := c.RealIP()
	userAgent := c.Request().UserAgent()

	response, err := lc.authService.Login(req, clientIP, userAgent)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	if !response.Success {
		return c.JSON(http.StatusUnauthorized, response)
	}

	return c.JSON(http.StatusOK, response)
}

func (lc *LoginController) Logout(c echo.Context) error {
	// Get token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authorization header required",
		})
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid authorization header format",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Get client IP and User Agent
	clientIP := c.RealIP()
	userAgent := c.Request().UserAgent()

	err := lc.authService.Logout(tokenString, clientIP, userAgent)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid token",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Logout successful",
	})
}

func (lc *LoginController) GetCurrentUser(c echo.Context) error {
	// Get token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authorization header required",
		})
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid authorization header format",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	user, err := lc.authService.GetCurrentUser(tokenString)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid token",
		})
	}

	return c.JSON(http.StatusOK, user)
}

func (lc *LoginController) RefreshToken(c echo.Context) error {
	// Get token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Authorization header required",
		})
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid authorization header format",
		})
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Get client IP and User Agent
	clientIP := c.RealIP()
	userAgent := c.Request().UserAgent()

	newToken, err := lc.authService.RefreshToken(tokenString, clientIP, userAgent)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid token",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"token":   newToken,
		"message": "Token refreshed successfully",
	})
}
