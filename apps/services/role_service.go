package services

import (
	"errors"
	"v01_system_backend/apps/models"
	"v01_system_backend/apps/repositories"
)

type RoleService struct {
	roleRepo *repositories.RoleRepository
}

func NewRoleService(roleRepo *repositories.RoleRepository) *RoleService {
	return &RoleService{roleRepo: roleRepo}
}

func (s *RoleService) GetAll(pagination *models.PaginationRequest) (*models.PaginationResponse, error) {
	pagination.SetDefaults()

	roles, totalRows, err := s.roleRepo.GetAll(pagination)
	if err != nil {
		return nil, err
	}

	// Get permissions for each role
	for i := range roles {
		permissions, _ := s.roleRepo.GetRolePermissions(roles[i].RolesID)
		roles[i].Permissions = permissions
	}

	totalPages := (totalRows + pagination.PageSize - 1) / pagination.PageSize

	return &models.PaginationResponse{
		Data:       roles,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}, nil
}

func (s *RoleService) GetByID(id int) (*models.Role, error) {
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, errors.New("role not found")
	}

	// Get role permissions
	permissions, _ := s.roleRepo.GetRolePermissions(role.RolesID)
	role.Permissions = permissions

	// Get role menus
	menus, _ := s.roleRepo.GetRoleMenus(role.RolesID)
	role.Menus = menus

	return role, nil
}

func (s *RoleService) Create(req *models.RoleCreateRequest, createdBy int) (*models.Role, error) {
	// Check if role name already exists
	existingRole, err := s.roleRepo.GetByName(req.RolesName)
	if err != nil {
		return nil, err
	}
	if existingRole != nil {
		return nil, errors.New("role name already exists")
	}

	// Check if role code already exists
	existingCode, err := s.roleRepo.GetByCode(req.RolesCode)
	if err != nil {
		return nil, err
	}
	if existingCode != nil {
		return nil, errors.New("role code already exists")
	}

	return s.roleRepo.Create(req, createdBy)
}

func (s *RoleService) Update(id int, req *models.RoleUpdateRequest, updatedBy int) (*models.Role, error) {
	// Check if role exists
	existingRole, err := s.roleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existingRole == nil {
		return nil, errors.New("role not found")
	}

	// Check if it's a system role
	if existingRole.IsSystemRole {
		return nil, errors.New("cannot update system role")
	}

	// Check if role name already exists (excluding current role)
	roleByName, err := s.roleRepo.GetByName(req.RolesName)
	if err != nil {
		return nil, err
	}
	if roleByName != nil && roleByName.RolesID != id {
		return nil, errors.New("role name already exists")
	}

	// Check if role code already exists (excluding current role)
	roleByCode, err := s.roleRepo.GetByCode(req.RolesCode)
	if err != nil {
		return nil, err
	}
	if roleByCode != nil && roleByCode.RolesID != id {
		return nil, errors.New("role code already exists")
	}

	return s.roleRepo.Update(id, req, updatedBy)
}

func (s *RoleService) Delete(id int, deletedBy int) error {
	// Check if role exists
	role, err := s.roleRepo.GetByID(id)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	// Check if it's a system role
	if role.IsSystemRole {
		return errors.New("cannot delete system role")
	}

	// Check if role is assigned to users
	userCount, err := s.roleRepo.GetUserCountByRole(id)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return errors.New("cannot delete role that is assigned to users")
	}

	return s.roleRepo.Delete(id, deletedBy)
}

func (s *RoleService) AssignPermissions(roleID int, permissionIDs []int, assignedBy int) error {
	// Check if role exists
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	return s.roleRepo.AssignPermissions(roleID, permissionIDs, assignedBy)
}

func (s *RoleService) RemovePermissions(roleID int, permissionIDs []int) error {
	// Check if role exists
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	return s.roleRepo.RemovePermissions(roleID, permissionIDs)
}

func (s *RoleService) AssignMenus(roleID int, menuPermissions []models.RoleMenuRequest, assignedBy int) error {
	// Check if role exists
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	return s.roleRepo.AssignMenus(roleID, menuPermissions, assignedBy)
}

func (s *RoleService) GetRolePermissions(roleID int) ([]*models.Permission, error) {
	permissions, err := s.roleRepo.GetRolePermissions(roleID)
	if err != nil {
		return nil, err
	}
	permPtrs := make([]*models.Permission, len(permissions))
	for i := range permissions {
		permPtrs[i] = &permissions[i]
	}
	return permPtrs, nil
}

func (s *RoleService) GetRoleMenus(roleID int) ([]*models.RoleMenu, error) {
	menuAccesses, err := s.roleRepo.GetRoleMenus(roleID)
	if err != nil {
		return nil, err
	}
	roleMenus := make([]*models.RoleMenu, len(menuAccesses))
	for i := range menuAccesses {
		roleMenus[i] = &models.RoleMenu{
			// Map fields from menuAccesses[i] to RoleMenu as needed
			// Example:
			// MenuID: menuAccesses[i].MenuID,
			// AccessLevel: menuAccesses[i].AccessLevel,
		}
	}
	return roleMenus, nil
}

// Add this method to your existing RoleService
func (s *RoleService) GetAllPermissions() ([]models.Permission, error) {
	// This should call the repository method to get all permissions
	return s.roleRepo.GetAllPermissions()
}
