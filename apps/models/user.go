package models

import (
	"time"
)

type User struct {
	UserAppsID          int        `json:"user_apps_id" db:"user_apps_id"`
	Username            string     `json:"username" db:"username"`
	Email               string     `json:"email" db:"email"`
	PasswordHash        string     `json:"-" db:"password_hash"`
	FirstName           string     `json:"first_name" db:"first_name"`
	LastName            string     `json:"last_name" db:"last_name"`
	StatusID            int        `json:"status_id" db:"status_id"`
	DepartmentID        *int       `json:"department_id" db:"department_id"`
	EmployeeID          *string    `json:"employee_id" db:"employee_id"`
	Phone               *string    `json:"phone" db:"phone"`
	AvatarURL           *string    `json:"avatar_url" db:"avatar_url"`
	LastLoginAt         *time.Time `json:"last_login_at" db:"last_login_at"`
	PasswordChangedAt   time.Time  `json:"password_changed_at" db:"password_changed_at"`
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until" db:"locked_until"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	BaseModel

	// Relations
	StatusName     *string `json:"status_name" db:"status_name"`
	DepartmentName *string `json:"department_name" db:"department_name"`
	Roles          []Role  `json:"roles,omitempty"`
}

type UserStatus struct {
	UsersApplicationStatusID int    `json:"users_application_status_id" db:"users_application_status_id"`
	StatusCode               string `json:"status_code" db:"status_code"`
	StatusName               string `json:"status_name" db:"status_name"`
	Description              string `json:"description" db:"description"`
	IsActive                 bool   `json:"is_active" db:"is_active"`
	BaseModel
}

type UserCreateRequest struct {
	Username     string `json:"username" validate:"required,min=3,max=50"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
	FirstName    string `json:"first_name" validate:"required,max=50"`
	LastName     string `json:"last_name" validate:"required,max=50"`
	StatusID     int    `json:"status_id" validate:"required"`
	DepartmentID *int   `json:"department_id"`
	EmployeeID   string `json:"employee_id" validate:"max=20"`
	Phone        string `json:"phone" validate:"max=20"`
	RoleIDs      []int  `json:"role_ids"`
}

type UserUpdateRequest struct {
	Username     string `json:"username" validate:"required,min=3,max=50"`
	Email        string `json:"email" validate:"required,email"`
	FirstName    string `json:"first_name" validate:"required,max=50"`
	LastName     string `json:"last_name" validate:"required,max=50"`
	StatusID     int    `json:"status_id" validate:"required"`
	DepartmentID *int   `json:"department_id"`
	EmployeeID   string `json:"employee_id" validate:"max=20"`
	Phone        string `json:"phone" validate:"max=20"`
	RoleIDs      []int  `json:"role_ids"`
}
