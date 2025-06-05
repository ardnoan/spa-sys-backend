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

type UserSession struct {
	SessionID    int        `json:"session_id" db:"session_id"`
	UserID       int        `json:"user_id" db:"user_id"`
	SessionToken string     `json:"session_token" db:"session_token"`
	IPAddress    *string    `json:"ip_address" db:"ip_address"`
	UserAgent    *string    `json:"user_agent" db:"user_agent"`
	LoginAt      time.Time  `json:"login_at" db:"login_at"`
	LogoutAt     *time.Time `json:"logout_at" db:"logout_at"`
	ExpiresAt    time.Time  `json:"expires_at" db:"expires_at"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
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

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User         User   `json:"user"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
