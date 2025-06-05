package services

import (
	"errors"
	"time"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/repositories"
	"v01_system_backend/apps/utils"
	"v01_system_backend/config"
)

type AuthService struct {
	userRepo *repositories.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo *repositories.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Login(req *models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, errors.New("account is locked due to multiple failed login attempts")
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		// Increment failed attempts
		s.userRepo.IncrementFailedAttempts(user.UserAppsID)

		// Lock account if max attempts reached
		maxAttempts := 5 // Should come from system settings
		if user.FailedLoginAttempts+1 >= maxAttempts {
			s.userRepo.LockUser(user.UserAppsID, 30) // Lock for 30 minutes
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
		return nil, err
	}

	// Generate refresh token (longer expiry)
	refreshToken, err := utils.GenerateToken(
		user.UserAppsID,
		user.Username,
		user.Email,
		s.cfg.JWT.Secret,
		s.cfg.JWT.Expire*7, // 7 times longer
	)
	if err != nil {
		return nil, err
	}

	// Update last login
	s.userRepo.UpdateLastLogin(user.UserAppsID)

	// Get user roles
	roles, _ := s.userRepo.GetUserRoles(user.UserAppsID)
	user.Roles = roles

	// Clear sensitive data
	user.PasswordHash = ""

	return &models.LoginResponse{
		User:         *user,
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(time.Hour * time.Duration(s.cfg.JWT.Expire)).Unix(),
	}, nil
}

func (s *AuthService) ChangePassword(userID int, req *models.ChangePasswordRequest) error {
	// Get current user
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

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	return s.userRepo.UpdatePassword(userID, hashedPassword, userID)
}

func (s *AuthService) ValidateToken(tokenString string) (*utils.JWTClaims, error) {
	return utils.ValidateToken(tokenString, s.cfg.JWT.Secret)
}
