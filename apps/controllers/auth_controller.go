package controllers

import (
	"v01_system_backend/apps/services"

	"github.com/labstack/echo/v4"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (ac *AuthController) Login(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Login endpoint"})
}

func (ac *AuthController) Logout(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "Logout endpoint"})
}

func (ac *AuthController) RefreshToken(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "RefreshToken endpoint"})
}

func (ac *AuthController) ChangePassword(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "ChangePassword endpoint"})
}

func (ac *AuthController) GetProfile(c echo.Context) error {
	return c.JSON(200, map[string]string{"message": "GetProfile endpoint"})
}
