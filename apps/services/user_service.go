// services/user_service.go
package services

import (
	"errors"
	"fmt"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/repositories"
	"v01_system_backend/apps/utils"
)

type UserService struct {
	userRepo     *repositories.UserRepository
	activityRepo *repositories.ActivityRepository
}

func NewUserService(userRepo *repositories.UserRepository, activityRepo *repositories.ActivityRepository) *UserService {
	return &UserService{
		userRepo:     userRepo,
		activityRepo: activityRepo,
	}
}

// In user_service.go, update GetUserRoles method:
func (s *UserService) GetUserRoles(userID int) ([]models.Role, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return s.userRepo.GetUserRoles(userID)
}

func (s *UserService) GetAll(pagination *models.PaginationRequest, filters *models.UserFilters) (*models.PaginationResponse, error) {
	pagination.SetDefaults()

	users, totalRows, err := s.userRepo.GetAll(pagination, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Get roles for each user
	for i := range users {
		roles, err := s.userRepo.GetUserRoles(users[i].UserAppsID)
		if err != nil {
			// Log error but continue
			continue
		}
		users[i].Roles = roles
	}

	totalPages := (totalRows + pagination.PageSize - 1) / pagination.PageSize

	return &models.PaginationResponse{
		Data:       users,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}, nil
}

func (s *UserService) GetByID(id int) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Get user roles
	roles, err := s.userRepo.GetUserRoles(user.UserAppsID)
	if err == nil {
		user.Roles = roles
	}

	return user, nil
}

func (s *UserService) Create(req *models.UserCreateRequest, createdBy int) (*models.User, error) {
	// Validate request
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check duplicates
	if err := s.checkDuplicateUser(req.Username, req.Email, 0); err != nil {
		return nil, err
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user, err := s.userRepo.Create(req, hashedPassword, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default roles if specified
	if len(req.RoleIDs) > 0 {
		if err := s.userRepo.AssignRoles(user.UserAppsID, req.RoleIDs, createdBy); err != nil {
			// Log error but don't fail the creation
		}
	}

	// Get user with roles
	return s.GetByID(user.UserAppsID)
}

func (s *UserService) Update(id int, req *models.UserUpdateRequest, updatedBy int) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Validate request
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Check if user exists
	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, errors.New("user not found")
	}

	// Check duplicates (excluding current user)
	if err := s.checkDuplicateUser(req.Username, req.Email, id); err != nil {
		return nil, err
	}

	// Update user
	user, err := s.userRepo.Update(id, req, updatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Get updated user with roles
	return s.GetByID(user.UserAppsID)
}

func (s *UserService) Delete(id int, deletedBy int) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Prevent self-deletion
	if id == deletedBy {
		return errors.New("cannot delete your own account")
	}

	return s.userRepo.Delete(id, deletedBy)
}

func (s *UserService) UpdateStatus(id int, status string, updatedBy int) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}

	// Validate status
	if !utils.IsValidStatus(status) {
		return errors.New("invalid status")
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.userRepo.UpdateStatus(id, status, updatedBy)
}

func (s *UserService) ResetPassword(id int, newPassword string, updatedBy int) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}

	// Validate password
	if err := utils.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return s.userRepo.UpdatePassword(id, hashedPassword, updatedBy)
}

// User Roles Management
func (s *UserService) GetUserRoles(userID int) ([]*models.Role, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return s.userRepo.GetUserRoles(userID)
}

func (s *UserService) AssignRoles(userID int, roleIDs []int, assignedBy int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if len(roleIDs) == 0 {
		return errors.New("no roles specified")
	}

	// Check if user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.userRepo.AssignRoles(userID, roleIDs, assignedBy)
}

func (s *UserService) RemoveRoles(userID int, roleIDs []int, removedBy int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if len(roleIDs) == 0 {
		return errors.New("no roles specified")
	}

	return s.userRepo.RemoveRoles(userID, roleIDs, removedBy)
}

// User Permissions
func (s *UserService) GetUserPermissions(userID int) ([]*models.Permission, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return s.userRepo.GetUserPermissions(userID)
}

// User Activities
func (s *UserService) GetUserActivities(userID int, pagination *models.PaginationRequest) (*models.PaginationResponse, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	pagination.SetDefaults()
	return s.activityRepo.GetUserActivities(userID, pagination)
}

// Private helper methods
func (s *UserService) validateCreateRequest(req *models.UserCreateRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if !utils.IsValidEmail(req.Email) {
		return errors.New("invalid email format")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		return err
	}
	return nil
}

func (s *UserService) validateUpdateRequest(req *models.UserUpdateRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if !utils.IsValidEmail(req.Email) {
		return errors.New("invalid email format")
	}
	return nil
}

func (s *UserService) checkDuplicateUser(username, email string, excludeID int) error {
	// Check username
	existingUser, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil && existingUser.UserAppsID != excludeID {
		return errors.New("username already exists")
	}

	// Check email
	existingEmail, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	}
	if existingEmail != nil && existingEmail.UserAppsID != excludeID {
		return errors.New("email already exists")
	}

	return nil
}
