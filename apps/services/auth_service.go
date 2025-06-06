package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/repositories"
	"v01_system_backend/apps/utils"
	"v01_system_backend/config"
)

type AuthService struct {
	userRepo    *repositories.UserRepository
	sessionRepo *repositories.SessionRepository
	cfg         *config.Config
}

func NewAuthService(userRepo *repositories.UserRepository, sessionRepo *repositories.SessionRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		cfg:         cfg,
	}
}

// Login authenticates user and returns tokens
func (s *AuthService) Login(req *models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		lockDuration := time.Until(*user.LockedUntil)
		return nil, fmt.Errorf("account is locked for %v due to multiple failed login attempts", lockDuration.Round(time.Minute))
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		// Increment failed attempts
		s.userRepo.IncrementFailedAttempts(user.UserAppsID)

		// Lock account if max attempts reached
		maxAttempts := s.cfg.Security.MaxLoginAttempts
		if maxAttempts == 0 {
			maxAttempts = 5 // default
		}

		if user.FailedLoginAttempts+1 >= maxAttempts {
			lockDuration := s.cfg.Security.LockDurationMinutes
			if lockDuration == 0 {
				lockDuration = 30 // default 30 minutes
			}
			s.userRepo.LockUser(user.UserAppsID, lockDuration)
		}

		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(
		user.UserAppsID,
		user.Username,
		user.Email,
		s.cfg.JWT.Secret,
		s.cfg.JWT.Expire,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := utils.GenerateToken(
		user.UserAppsID,
		user.Username,
		user.Email,
		s.cfg.JWT.RefreshSecret,
		s.cfg.JWT.RefreshExpire,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	session := &models.UserSession{
		UserID:       user.UserAppsID,
		SessionToken: refreshToken,
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
		LoginAt:      time.Now(),
		ExpiresAt:    time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.RefreshExpire)),
		IsActive:     true,
	}

	if err := s.sessionRepo.Create(session); err != nil {
		// Log error but don't fail login
		fmt.Printf("Warning: Failed to create session: %v\n", err)
	}

	// Update last login and reset failed attempts
	s.userRepo.UpdateLastLogin(user.UserAppsID)

	// Get user profile
	profile, err := s.buildUserProfile(user)
	if err != nil {
		return nil, fmt.Errorf("failed to build user profile: %w", err)
	}

	return &models.LoginResponse{
		User:         *profile,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.Expire)).Unix(),
		TokenType:    "Bearer",
	}, nil
}

// RefreshToken generates new access token using refresh token
func (s *AuthService) RefreshToken(refreshTokenString string) (*models.RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := utils.ValidateToken(refreshTokenString, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if session exists and is active
	session, err := s.sessionRepo.GetByToken(refreshTokenString)
	if err != nil || session == nil || !session.IsActive {
		return nil, errors.New("invalid or expired refresh token")
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("refresh token expired")
	}

	// Generate new access token
	token, err := utils.GenerateToken(
		claims.UserID,
		claims.Username,
		claims.Email,
		s.cfg.JWT.Secret,
		s.cfg.JWT.Expire,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token: %w", err)
	}

	// Generate new refresh token
	newRefreshToken, err := utils.GenerateToken(
		claims.UserID,
		claims.Username,
		claims.Email,
		s.cfg.JWT.RefreshSecret,
		s.cfg.JWT.RefreshExpire,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new refresh token: %w", err)
	}

	// Update session with new refresh token
	session.SessionToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.RefreshExpire))
	s.sessionRepo.Update(session)

	return &models.RefreshTokenResponse{
		Token:        token,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.Expire)).Unix(),
		TokenType:    "Bearer",
	}, nil
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(userID int, req *models.ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	// Get current password hash
	currentUser, err := s.userRepo.GetByUsername(user.Username)
	if err != nil {
		return err
	}

	// Verify current password
	if !utils.CheckPassword(req.CurrentPassword, currentUser.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	// Validate new password
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	return s.userRepo.UpdatePassword(userID, hashedPassword, userID)
}

// GetProfile returns user profile
func (s *AuthService) GetProfile(userID int) (*models.UserProfile, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return s.buildUserProfile(user)
}

// UpdateProfile updates user profile
func (s *AuthService) UpdateProfile(userID int, req *models.UpdateProfileRequest) (*models.UserProfile, error) {
	// Get current user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Validate phone if provided
	if req.Phone != "" {
		if err := utils.ValidatePhone(req.Phone); err != nil {
			return nil, err
		}
	}

	// Update profile
	updateReq := &models.UserUpdateRequest{
		Username:     user.Username, // Keep existing username
		Email:        user.Email,    // Keep existing email
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		StatusID:     user.StatusID,     // Keep existing status
		DepartmentID: user.DepartmentID, // Keep existing department
		EmployeeID:   *user.EmployeeID,  // Keep existing employee ID
		Phone:        req.Phone,
	}

	updatedUser, err := s.userRepo.Update(userID, updateReq, userID)
	if err != nil {
		return nil, err
	}

	return s.buildUserProfile(updatedUser)
}

// ForgotPassword initiates password reset process
func (s *AuthService) ForgotPassword(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return err
	}

	if user == nil {
		// Don't reveal that email doesn't exist
		return nil
	}

	// Generate reset token
	resetToken, err := s.generateResetToken()
	if err != nil {
		return err
	}

	// Save reset token (implement repository method)
	// s.userRepo.SavePasswordResetToken(user.UserAppsID, resetToken)

	// Send email (implement email service)
	// s.emailService.SendPasswordResetEmail(email, resetToken)

	return nil
}

// ResetPassword resets password using reset token
func (s *AuthService) ResetPassword(req *models.ResetPasswordRequest) error {
	// Validate reset token (implement)
	// userID, err := s.userRepo.ValidatePasswordResetToken(req.Token)
	// if err != nil {
	//     return errors.New("invalid or expired reset token")
	// }

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password (implement userID retrieval)
	// return s.userRepo.UpdatePassword(userID, hashedPassword, userID)

	return errors.New("password reset not implemented yet")
}

// ValidateToken validates JWT token
func (s *AuthService) ValidateToken(tokenString string) (*utils.JWTClaims, error) {
	return utils.ValidateToken(tokenString, s.cfg.JWT.Secret)
}

// Helper methods
func (s *AuthService) buildUserProfile(user *models.User) (*models.UserProfile, error) {
	// Get user roles
	roles, err := s.userRepo.GetUserRoles(user.UserAppsID)
	if err != nil {
		return nil, err
	}

	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.RolesName
	}

	// Get user permissions (implement if needed)
	permissions := []string{} // TODO: implement permission retrieval

	return &models.UserProfile{
		UserID:      user.UserAppsID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Phone:       user.Phone,
		AvatarURL:   user.AvatarURL,
		Status:      *user.StatusName,
		Department:  user.DepartmentName,
		LastLoginAt: user.LastLoginAt,
		Roles:       roleNames,
		Permissions: permissions,
	}, nil
}

func (s *AuthService) generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
