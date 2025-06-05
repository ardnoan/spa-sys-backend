package handlers

import (
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/services"
	"v01_system_backend/apps/utils"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Validate request
	if req.Username == "" || req.Password == "" {
		return utils.BadRequestResponse(c, "Username and password are required", nil)
	}

	// Get client info
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	// Login
	response, err := h.authService.Login(&req, ipAddress, userAgent)
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, "Login successful", response)
}

func (h *AuthHandler) ChangePassword(c echo.Context) error {
	var req models.ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request data", err.Error())
	}

	// Validate request
	if req.CurrentPassword == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
		return utils.BadRequestResponse(c, "All password fields are required", nil)
	}

	if req.NewPassword != req.ConfirmPassword {
		return utils.BadRequestResponse(c, "New password and confirm password do not match", nil)
	}

	// Get user ID from context
	userID := c.Get("user_id").(int)

	// Change password
	if err := h.authService.ChangePassword(userID, &req); err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Password changed successfully", nil)
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	// This would be implemented to get current user profile
	userID := c.Get("user_id").(int)
	username := c.Get("username").(string)
	email := c.Get("email").(string)

	profile := map[string]interface{}{
		"user_id":  userID,
		"username": username,
		"email":    email,
	}

	return utils.SuccessResponse(c, "Profile retrieved successfully", profile)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	// For now, just return success
	// In a complete implementation, you would invalidate the token
	return utils.SuccessResponse(c, "Logged out successfully", nil)
}
