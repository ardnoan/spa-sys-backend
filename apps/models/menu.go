package models

import "time"

type Menu struct {
	MenusID   int     `json:"menus_id" db:"menus_id"`
	MenuCode  string  `json:"menu_code" db:"menu_code"`
	MenuName  string  `json:"menu_name" db:"menu_name"`
	ParentID  *int    `json:"parent_id" db:"parent_id"`
	IconName  *string `json:"icon_name" db:"icon_name"`
	Route     *string `json:"route" db:"route"`
	MenuOrder int     `json:"menu_order" db:"menu_order"`
	IsVisible bool    `json:"is_visible" db:"is_visible"`
	IsActive  bool    `json:"is_active" db:"is_active"`
	BaseModel

	// Relations
	Children []Menu      `json:"children,omitempty"`
	Access   *MenuAccess `json:"access,omitempty"`
}

type MenuAccess struct {
	RoleMenuID  int       `json:"role_menu_id" db:"role_menu_id"`
	RoleID      int       `json:"role_id" db:"role_id"`
	MenuID      int       `json:"menu_id" db:"menu_id"`
	CanView     bool      `json:"can_view" db:"can_view"`
	CanCreate   bool      `json:"can_create" db:"can_create"`
	CanModify   bool      `json:"can_modify" db:"can_modify"`
	CanDelete   bool      `json:"can_delete" db:"can_delete"`
	CanUpload   bool      `json:"can_upload" db:"can_upload"`
	CanDownload bool      `json:"can_download" db:"can_download"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	CreatedBy   *int      `json:"created_by" db:"created_by"`
}

type UserMenuResponse struct {
	Menus []Menu `json:"menus"`
}
