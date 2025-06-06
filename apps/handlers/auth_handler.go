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

// Login handles user authentication
func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest

	// Bind request data
	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request format", err.Error())
	}

	// Validate input
	if err := utils.ValidateStruct(&req); err != nil {
		validationErrors := utils.GetValidationErrors(err)
		return utils.BadRequestResponse(c, "Validation failed", validationErrors)
	}

	// Basic validation
	if req.Username == "" || req.Password == "" {
		return utils.BadRequestResponse(c, "Username and password are required", nil)
	}

	// Get client info
	ipAddress := c.RealIP()
	userAgent := c.Request().UserAgent()

	// Process login
	response, err := h.authService.Login(&req, ipAddress, userAgent)
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, "Login successful", response)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c echo.Context) error {
	// Get token from header
	token := c.Request().Header.Get("Authorization")
	if token != "" {
		// Remove "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Optional: Add token to blacklist (implement if needed)
		// h.authService.BlacklistToken(token)
	}

	return utils.SuccessResponse(c, "Logged out successfully", nil)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req models.RefreshTokenRequest

	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request format", err.Error())
	}

	if req.RefreshToken == "" {
		return utils.BadRequestResponse(c, "Refresh token is required", nil)
	}

	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return utils.UnauthorizedResponse(c, err.Error())
	}

	return utils.SuccessResponse(c, "Token refreshed successfully", response)
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	var req models.ChangePasswordRequest

	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request format", err.Error())
	}

	// Validate input
	if err := utils.ValidateStruct(&req); err != nil {
		validationErrors := utils.GetValidationErrors(err)
		return utils.BadRequestResponse(c, "Validation failed", validationErrors)
	}

	// Get user ID from JWT context
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return utils.UnauthorizedResponse(c, "Invalid user session")
	}

	// Change password
	if err := h.authService.ChangePassword(userID, &req); err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Password changed successfully", nil)
}

// GetProfile returns current user profile
func (h *AuthHandler) GetProfile(c echo.Context) error {
	// Get user info from JWT context
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return utils.UnauthorizedResponse(c, "Invalid user session")
	}

	profile, err := h.authService.GetProfile(userID)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get profile", err.Error())
	}

	return utils.SuccessResponse(c, "Profile retrieved successfully", profile)
}

// GetMe returns current user info (alias for GetProfile)
func (h *AuthHandler) GetMe(c echo.Context) error {
	return h.GetProfile(c)
}

// UpdateProfile updates current user profile
func (h *AuthHandler) UpdateProfile(c echo.Context) error {
	var req models.UpdateProfileRequest

	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request format", err.Error())
	}

	// Validate input
	if err := utils.ValidateStruct(&req); err != nil {
		validationErrors := utils.GetValidationErrors(err)
		return utils.BadRequestResponse(c, "Validation failed", validationErrors)
	}

	// Get user ID from JWT context
	userID, ok := c.Get("user_id").(int)
	if !ok {
		return utils.UnauthorizedResponse(c, "Invalid user session")
	}

	profile, err := h.authService.UpdateProfile(userID, &req)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Profile updated successfully", profile)
}

// ForgotPassword handles password reset request
func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req models.ForgotPasswordRequest

	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request format", err.Error())
	}

	if req.Email == "" {
		return utils.BadRequestResponse(c, "Email is required", nil)
	}

	if err := utils.ValidateEmail(req.Email); err != nil {
		return utils.BadRequestResponse(c, "Invalid email format", nil)
	}

	err := h.authService.ForgotPassword(req.Email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return utils.SuccessResponse(c, "If the email exists, password reset instructions have been sent", nil)
	}

	return utils.SuccessResponse(c, "Password reset instructions have been sent to your email", nil)
}

// ResetPassword handles password reset with token
func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req models.ResetPasswordRequest

	if err := c.Bind(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request format", err.Error())
	}

	// Validate input
	if err := utils.ValidateStruct(&req); err != nil {
		validationErrors := utils.GetValidationErrors(err)
		return utils.BadRequestResponse(c, "Validation failed", validationErrors)
	}

	err := h.authService.ResetPassword(&req)
	if err != nil {
		return utils.BadRequestResponse(c, err.Error(), nil)
	}

	return utils.SuccessResponse(c, "Password reset successfully", nil)
}
