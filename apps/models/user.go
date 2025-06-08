// models/user.go
package models

import (
	"time"
)

type User struct {
	UserAppsID  int        `json:"user_apps_id" db:"user_apps_id"`
	Username    string     `json:"username" db:"username"`
	Email       string     `json:"email" db:"email"`
	FullName    string     `json:"full_name" db:"full_name"`
	Status      string     `json:"status" db:"status"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy   int        `json:"created_by" db:"created_by"`
	UpdatedBy   *int       `json:"updated_by" db:"updated_by"`
	Roles       []*Role    `json:"roles,omitempty"`
}

type UserCreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8"`
	Status   string `json:"status" validate:"omitempty,oneof=active inactive"`
	RoleIDs  []int  `json:"role_ids,omitempty"`
}

type UserUpdateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Status   string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type StatusUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended"`
}

type PasswordResetRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type RoleAssignmentRequest struct {
	RoleIDs []int `json:"role_ids" validate:"required,min=1"`
}

type RoleRemovalRequest struct {
	RoleIDs []int `json:"role_ids" validate:"required,min=1"`
}

type UserFilters struct {
	Status string `json:"status"`
	Search string `json:"search"`
	Role   string `json:"role"`
}
