package services

import (
	"errors"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/repositories"
	"v01_system_backend/apps/utils"
)

type UserService struct {
	userRepo *repositories.UserRepository
}

func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetAll(pagination *models.PaginationRequest) (*models.PaginationResponse, error) {
	pagination.SetDefaults()

	users, totalRows, err := s.userRepo.GetAll(pagination)
	if err != nil {
		return nil, err
	}

	// Get roles for each user
	for i := range users {
		roles, _ := s.userRepo.GetUserRoles(users[i].UserAppsID)
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
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	// Get user roles
	roles, _ := s.userRepo.GetUserRoles(user.UserAppsID)
	user.Roles = roles

	return user, nil
}

func (s *UserService) Create(req *models.UserCreateRequest, createdBy int) (*models.User, error) {
	// Check if username already exists
	existingUser, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	existingEmail, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingEmail != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := s.userRepo.Create(req, hashedPassword, createdBy)
	if err != nil {
		return nil, err
	}

	// Get user roles
	roles, _ := s.userRepo.GetUserRoles(user.UserAppsID)
	user.Roles = roles

	return user, nil
}

func (s *UserService) Update(id int, req *models.UserUpdateRequest, updatedBy int) (*models.User, error) {
	// Check if user exists
	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existingUser == nil {
		return nil, errors.New("user not found")
	}

	// Check if username already exists (excluding current user)
	userByUsername, err := s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if userByUsername != nil && userByUsername.UserAppsID != id {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists (excluding current user)
	userByEmail, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if userByEmail != nil && userByEmail.UserAppsID != id {
		return nil, errors.New("email already exists")
	}

	// Update user
	user, err := s.userRepo.Update(id, req, updatedBy)
	if err != nil {
		return nil, err
	}

	// Get user roles
	roles, _ := s.userRepo.GetUserRoles(user.UserAppsID)
	user.Roles = roles

	return user, nil
}

func (s *UserService) Delete(id int, deletedBy int) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.userRepo.Delete(id, deletedBy)
}

func (s *UserService) ResetPassword(id int, newPassword string, updatedBy int) error {
	// Check if user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(id, hashedPassword, updatedBy)
}
