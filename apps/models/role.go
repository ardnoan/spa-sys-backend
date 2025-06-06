package models

import "time"

type Role struct {
	RolesID      int    `json:"roles_id" db:"roles_id"`
	RolesName    string `json:"roles_name" db:"roles_name"`
	RolesCode    string `json:"roles_code" db:"roles_code"`
	Description  string `json:"description" db:"description"`
	IsSystemRole bool   `json:"is_system_role" db:"is_system_role"`
	IsActive     bool   `json:"is_active" db:"is_active"`
	BaseModel

	// Relations
	Permissions []Permission `json:"permissions,omitempty"`
	Menus       []MenuAccess `json:"menus,omitempty"`
}

type RolePermission struct {
	RolePermissionID int       `json:"role_permission_id" db:"role_permission_id"`
	RoleID           int       `json:"role_id" db:"role_id"`
	PermissionID     int       `json:"permission_id" db:"permission_id"`
	GrantedAt        time.Time `json:"granted_at" db:"granted_at"`
	GrantedBy        *int      `json:"granted_by" db:"granted_by"`
	IsActive         bool      `json:"is_active" db:"is_active"`
}

type UserRole struct {
	UserRoleID int       `json:"user_role_id" db:"user_role_id"`
	UserID     int       `json:"user_id" db:"user_id"`
	RoleID     int       `json:"role_id" db:"role_id"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
	AssignedBy *int      `json:"assigned_by" db:"assigned_by"`
	IsActive   bool      `json:"is_active" db:"is_active"`
}

// RoleCreateRequest for creating new roles
type RoleCreateRequest struct {
	RolesName     string  `json:"roles_name" validate:"required,max=50"`
	RolesCode     string  `json:"roles_code" validate:"required,max=20"`
	Description   *string `json:"description"`
	IsSystemRole  bool    `json:"is_system_role"`
	PermissionIDs []int   `json:"permission_ids"`
}

// RoleUpdateRequest for updating roles
type RoleUpdateRequest struct {
	RolesName     string  `json:"roles_name" validate:"required,max=50"`
	RolesCode     string  `json:"roles_code" validate:"required,max=20"`
	Description   *string `json:"description"`
	PermissionIDs []int   `json:"permission_ids"`
}

// RoleMenuRequest for assigning menus to roles
type RoleMenuRequest struct {
	MenuID      int  `json:"menu_id" validate:"required"`
	CanView     bool `json:"can_view"`
	CanCreate   bool `json:"can_create"`
	CanModify   bool `json:"can_modify"`
	CanDelete   bool `json:"can_delete"`
	CanUpload   bool `json:"can_upload"`
	CanDownload bool `json:"can_download"`
}

// RoleMenu represents the relationship between roles and menus
type RoleMenu struct {
	RoleMenuID  int  `json:"role_menu_id" db:"role_menu_id"`
	RoleID      int  `json:"role_id" db:"role_id"`
	MenuID      int  `json:"menu_id" db:"menu_id"`
	CanView     bool `json:"can_view" db:"can_view"`
	CanCreate   bool `json:"can_create" db:"can_create"`
	CanModify   bool `json:"can_modify" db:"can_modify"`
	CanDelete   bool `json:"can_delete" db:"can_delete"`
	CanUpload   bool `json:"can_upload" db:"can_upload"`
	CanDownload bool `json:"can_download" db:"can_download"`
	BaseModel
}
