// apps/services/user_service.go
package services

import (
	"errors"
	"fmt"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/repositories"
	"v01_system_backend/apps/utils"
)

// UserService menghandle business logic untuk operasi user
// Layer ini berada di antara controller dan repository
// Bertanggung jawab untuk:
// - Validasi business rules
// - Koordinasi antar repository
// - Data transformation
type UserService struct {
	userRepo     *repositories.UserRepository     // Repository untuk operasi user
	activityRepo *repositories.ActivityRepository // Repository untuk logging activity
}

// NewUserService constructor dengan dependency injection
func NewUserService(userRepo *repositories.UserRepository, activityRepo *repositories.ActivityRepository) *UserService {
	return &UserService{
		userRepo:     userRepo,
		activityRepo: activityRepo,
	}
}

// GetAll mengambil semua user dengan pagination dan filtering
// BUSINESS LOGIC:
// 1. Set default values untuk pagination
// 2. Query data dari repository dengan filter
// 3. Enrich data dengan user roles
// 4. Format response dengan pagination metadata
func (s *UserService) GetAll(pagination *models.PaginationRequest, filters *models.UserFilters) (*models.PaginationResponse, error) {
	// Set default values jika tidak ada di request
	// Default: page=1, page_size=10, sort_dir=ASC
	pagination.SetDefaults()

	// Query users dari database dengan pagination dan filter
	users, totalRows, err := s.userRepo.GetAll(pagination, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Enrich setiap user dengan data roles
	// Loop untuk menambahkan role information ke setiap user
	for i := range users {
		roles, err := s.userRepo.GetUserRoles(users[i].UserAppsID)
		if err != nil {
			// Log error tapi jangan stop process
			// Bisa pakai logger di sini untuk production
			continue
		}
		users[i].Roles = roles
	}

	// Hitung total pages untuk pagination
	totalPages := (totalRows + pagination.PageSize - 1) / pagination.PageSize

	// Return response dengan metadata pagination
	return &models.PaginationResponse{
		Data:       users,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}, nil
}

// GetByID mengambil user berdasarkan ID
// BUSINESS LOGIC:
// 1. Validasi ID input
// 2. Query dari repository
// 3. Handle case user not found
// 4. Enrich dengan role data
func (s *UserService) GetByID(id int) (*models.User, error) {
	// Validasi input ID
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Query user dari database
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Handle case user tidak ditemukan
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Enrich user dengan role data
	roles, err := s.userRepo.GetUserRoles(user.UserAppsID)
	if err == nil { // Hanya assign jika tidak ada error
		user.Roles = roles
	}

	return user, nil
}

// Create membuat user baru
// BUSINESS LOGIC:
// 1. Validasi input data
// 2. Check duplicate username/email
// 3. Hash password
// 4. Create user record
// 5. Assign default roles jika ada
func (s *UserService) Create(req *models.UserCreateRequest, createdBy int) (*models.User, error) {
	// Validasi business rules untuk create
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Check apakah username atau email sudah ada
	// excludeID = 0 karena ini create baru
	if err := s.checkDuplicateUser(req.Username, req.Email, 0); err != nil {
		return nil, err
	}

	// Hash password menggunakan bcrypt
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user record di database
	user, err := s.userRepo.Create(req, hashedPassword, createdBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Assign default roles jika ada dalam request
	if len(req.RoleIDs) > 0 {
		if err := s.userRepo.AssignRoles(user.UserAppsID, req.RoleIDs, createdBy); err != nil {
			// Log error tapi tidak gagalkan create user
			// Dalam production, bisa pakai logger
		}
	}

	// Return user dengan role data
	return s.GetByID(user.UserAppsID)
}

// Update mengupdate user existing
// BUSINESS LOGIC:
// 1. Validasi ID dan input
// 2. Check user exists
// 3. Check duplicate untuk username/email baru
// 4. Update record
func (s *UserService) Update(id int, req *models.UserUpdateRequest, updatedBy int) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user ID")
	}

	// Validasi input data
	if err := s.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	// Check apakah user yang akan diupdate ada
	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, errors.New("user not found")
	}

	// Check duplicate username/email (exclude current user)
	if err := s.checkDuplicateUser(req.Username, req.Email, id); err != nil {
		return nil, err
	}

	// Update user record
	user, err := s.userRepo.Update(id, req, updatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Return updated user dengan role data
	return s.GetByID(user.UserAppsID)
}

// Delete melakukan soft delete user
// BUSINESS LOGIC:
// 1. Validasi ID
// 2. Check user exists
// 3. Prevent self deletion
// 4. Soft delete (set is_active = false)
func (s *UserService) Delete(id int, deletedBy int) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}

	// Check user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Prevent user menghapus dirinya sendiri
	if id == deletedBy {
		return errors.New("cannot delete your own account")
	}

	// Soft delete user
	return s.userRepo.Delete(id, deletedBy)
}

// UpdateStatus mengupdate status user (active/inactive/suspended)
func (s *UserService) UpdateStatus(id int, status string, updatedBy int) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}

	// Validasi status value
	if !utils.IsValidStatus(status) {
		return errors.New("invalid status")
	}

	// Check user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.userRepo.UpdateStatus(id, status, updatedBy)
}

// ResetPassword mereset password user
// BUSINESS LOGIC:
// 1. Validasi password requirements
// 2. Hash password baru
// 3. Simpan ke password history
// 4. Update password
func (s *UserService) ResetPassword(id int, newPassword string, updatedBy int) error {
	if id <= 0 {
		return errors.New("invalid user ID")
	}

	// Validasi password strength
	if err := utils.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Check user exists
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Hash password baru
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password (repository akan handle password history)
	return s.userRepo.UpdatePassword(id, hashedPassword, updatedBy)
}

// === ROLE MANAGEMENT METHODS ===

// AssignRoles assign role ke user
func (s *UserService) AssignRoles(userID int, roleIDs []int, assignedBy int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if len(roleIDs) == 0 {
		return errors.New("no roles specified")
	}

	// Check user exists
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	// Repository akan handle transaction untuk assign roles
	return s.userRepo.AssignRoles(userID, roleIDs, assignedBy)
}

// RemoveRoles remove role dari user
func (s *UserService) RemoveRoles(userID int, roleIDs []int, removedBy int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}
	if len(roleIDs) == 0 {
		return errors.New("no roles specified")
	}

	return s.userRepo.RemoveRoles(userID, roleIDs, removedBy)
}

// GetUserPermissions mengambil semua permission user melalui roles
func (s *UserService) GetUserPermissions(userID int) ([]*models.Permission, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	return s.userRepo.GetUserPermissions(userID)
}

// === HELPER METHODS ===

// validateCreateRequest validasi data untuk create user
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

// validateUpdateRequest validasi data untuk update user
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

// checkDuplicateUser check duplicate username atau email
// excludeID untuk mengecualikan user tertentu (untuk update)
func (s *UserService) checkDuplicateUser(username, email string, excludeID int) error {
	// Check username duplicate
	existingUser, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil && existingUser.UserAppsID != excludeID {
		return errors.New("username already exists")
	}

	// Check email duplicate
	existingEmail, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return fmt.Errorf("failed to check email: %w", err)
	}
	if existingEmail != nil && existingEmail.UserAppsID != excludeID {
		return errors.New("email already exists")
	}

	return nil
}
