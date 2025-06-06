package models

type Permission struct {
	PermissionsID  int    `json:"permissions_id" db:"permissions_id"`
	PermissionCode string `json:"permission_code" db:"permission_code"`
	PermissionName string `json:"permission_name" db:"permission_name"`
	Description    string `json:"description" db:"description"`
	ModuleName     string `json:"module_name" db:"module_name"`
	IsActive       bool   `json:"is_active" db:"is_active"`
	BaseModel
}

type PermissionCreateRequest struct {
	PermissionCode string `json:"permission_code" validate:"required,max=50"`
	PermissionName string `json:"permission_name" validate:"required,max=100"`
	Description    string `json:"description"`
	ModuleName     string `json:"module_name" validate:"max=50"`
}

type PermissionUpdateRequest struct {
	PermissionCode string `json:"permission_code" validate:"required,max=50"`
	PermissionName string `json:"permission_name" validate:"required,max=100"`
	Description    string `json:"description"`
	ModuleName     string `json:"module_name" validate:"max=50"`
}
